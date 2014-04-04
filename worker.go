package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/iron-io/iron_go/worker"
)

type dotWorker struct {
	runtime string
	stack   string
	name    string
	remote  bool
	build   string // TODO(reed): ??? "npm install"?
	exec    string
	pip     map[string]string
	deb     []string
	jar     []string
	gems    map[string]string
	envs    map[string]string
	files   map[string]string
	dirs    map[string]string
}

func mergeFiles(source worker.CodeSource, files map[string]string) error {
	for f, storeAs := range files {
		contents, err := ioutil.ReadFile(f)
		if err != nil {
			return err
		}
		source[storeAs] = contents
	}
	return nil
}

func mergeDirs(source worker.CodeSource, dirs map[string]string) error {
	for dir, storeAs := range dirs {
		dir, err := filepath.Abs(dir)
		if err != nil {
			return err
		}
		err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}

			contents, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			path = path[len(dir):]
			source[storeAs+path] = contents
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// TODO(reed): merge gems
func mergeGems(source worker.CodeSource, gems map[string]string) error {
	for gem, version := range gems {
		out, _ := exec.Command("gem", "dependency", gem, "-v", version, "--pipe").Output()
		scanner := bufio.NewScanner(bytes.NewReader(out))
		for scanner.Scan() {
			//out, _ := exec.Command("gem", "contents", scanner.Text()).Output()
			//fmt.Println(string(out))
		}
	}
	return nil
}

func fixGithubURL(url string) string {
	if strings.HasPrefix(url, "http://github.com/") || strings.HasPrefix(url, "https://github.com/") {
		url = strings.Replace(url, "//github.com/", "//raw.github.com/", 1)
		url = strings.Replace(url, "/blob/", "/", 1)
	}
	return url
}

// TODO need to incorporate flags here for, e.g. max-concurrency, retries
func (dw *dotWorker) code() (worker.Code, error) {
	codes := worker.Code{
		Name:     dw.name,
		Runtime:  `sh`,
		FileName: `__runner__.sh`,
	}

	execContents, err := ioutil.ReadFile(dw.exec)
	if err != nil {
		return codes, err
	}
	execPath := filepath.Base(dw.exec)

	source := worker.CodeSource{
		dw.exec: execContents,
	}
	runtimeText, err := runtime(source, dw.runtime, execPath)
	if err != nil {
		return codes, err
	}

	runner := []byte(RUNNER +
		runtimeText + ` \"$@\"`) //+`#{File.basename(@exec.path)} #{params}

	source[`__runner__.sh`] = runner

	if err = mergeFiles(source, dw.files); err != nil {
		return codes, err
	}
	if err = mergeDirs(source, dw.dirs); err != nil {
		return codes, err
	}
	if err = mergeGems(source, dw.gems); err != nil {
		return codes, err
	}

	codes.Source = source

	return codes, nil
}

var RUNNER = `
#!/bin/sh
# TODO(reed): #{IronWorkerNG.full_version}

root() {
  while [ $# -gt 1 ]; do
  if [ "$1" = "-d" ]; then
  printf "%s" "$2"
  break
  fi

  shift
  done
}

cd "$(root "$@")"

LD_LIBRARY_PATH=.:./lib:./__debs__/usr/lib:./__debs__/usr/lib/x86_64-linux-gnu:./__debs__/lib:./__debs__/lib/x86_64-linux-gnu
export LD_LIBRARY_PATH

PATH=.:./bin:./__debs__/usr/bin:./__debs__/bin:$PATH
export PATH

# TODO(reed): #{container.runner_additions}

# TODO(reed): #{runtime_run_code(local, params)}
`

// name = someName.worker
//
// TODO(reed): the image viewer thing? uh wha?
func parseWorker(dotWorkerFile string) (*dotWorker, error) {
	if strings.HasPrefix(dotWorkerFile, "http://") || strings.HasPrefix(dotWorkerFile, "https://") {
		url := fixGithubURL(dotWorkerFile)
		return nil, errors.New("not ready for that yet hot shot") // TODO(reed): turnkey

		resp, _ := http.Get(url)
		fmt.Println(resp.Body)
		fmt.Println(filepath.Base(url))
	}

	dw := &dotWorker{
		name:  dotWorkerFile[:len(dotWorkerFile)-7], // TODO(reed): camel_case ?
		pip:   make(map[string]string),
		gems:  make(map[string]string),
		envs:  make(map[string]string),
		files: make(map[string]string),
		dirs:  make(map[string]string),
	}

	f, err := os.Open(dotWorkerFile)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() { // lines at a time
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") { // no comments, blank lines
			continue
		}
		// recursively descend lines with comma postfix
		for strings.HasSuffix(line, ",") {
			if !scanner.Scan() {
				return nil, errors.New("expected new line with value after: " + line)
			}
			line += " " + strings.TrimSpace(scanner.Text())
		}

		// at this point, words first token is the key and all comma separated
		// values are tokens after it, possibly 1 w/o comma
		err := dw.parseLine(line)
		if err != nil {
			return nil, err
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return dw, nil
}

func isQuote(r rune) bool {
	return r == '\'' || r == '"'
}

// TODO(reed): these are handled wrong and space separated would work fine
func isComma(r rune) bool {
	return r == ','
}

// ScanWords is a split function for a Scanner that returns each
// comma-separated word of text, with surrounding spaces and quotes deleted.
// Also trailing comments will be sliced off.
// TODO(reed): true/false for remote
func scanWords(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// Skip leading spaces.
	start := 0
	for width := 0; start < len(data); start += width {
		var r rune
		r, width = utf8.DecodeRune(data[start:])
		if r == '#' {
			start = len(data) // skip comments
		}
		if isQuote(r) { // at our word
			start += width
			break
		}

	}
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	// Scan until quote, marking end of word.
	for width, i := 0, start; i < len(data); i += width {
		var r rune
		r, width = utf8.DecodeRune(data[i:])
		if isQuote(r) {
			return i + width, data[start:i], nil
		}
	}
	// If we're at EOF, we have a final, non-empty, non-terminated word. Return it.
	if atEOF && len(data) > start {
		return len(data), data[start:], nil
	}
	// Request more data.
	return 0, nil, nil
}

// TODO(reed): yeahh these aren't guaranteed to be lines, so new name
// TODO(reed): scanner will also allow things like 'hello" or "hello'
// TODO(reed): small methods, bud
func (dw *dotWorker) parseLine(arg string) error {
	key := strings.Fields(arg)[0] // first field is space separated
	scanner := bufio.NewScanner(strings.NewReader(arg[len(key):]))
	scanner.Split(scanWords) // custom func above

	switch key {
	case "runtime":
		if !scanner.Scan() {
			return errors.New("runtime takes one arg")
		}
		runtime := scanner.Text()
		switch runtime {
		// TODO(reed): ehhh, duplication of concerns? server take care of this?
		case "binary", "go", "java", "mono", "node", "php", "python", "ruby", "perl":
			dw.runtime = runtime
		default:
			return errors.New(runtime + " not a valid runtime")
		}
	case "stack":
		if !scanner.Scan() {
			return errors.New("stack takes one arg")
		}
		// TODO(reed): validate stack here?
		dw.stack = scanner.Text()
	case "name":
		if !scanner.Scan() {
			return errors.New("name takes one arg")
		}
		dw.name = scanner.Text()
	case "set_env":
		if !scanner.Scan() {
			return errors.New("set_env takes two args: \"KEY\", \"VALUE\"")
		}
		key := scanner.Text()
		if !scanner.Scan() {
			return errors.New("set_env takes two args: \"KEY\", \"VALUE\"")
		}
		value := scanner.Text()
		dw.envs[key] = value
	case "full_remote_build", "remote":
		if !scanner.Scan() {
			dw.remote = true
			break
		}
		switch scanner.Text() {
		case "true":
			dw.remote = true
		case "false":
			dw.remote = false
		default:
			return errors.New("full_remote_build can only be true or false")
		}
	case "build":
		// TODO what?
	case "exec":
		if !scanner.Scan() {
			return errors.New("exec takes a file path")
		}
		dw.exec = scanner.Text()
		if scanner.Scan() {
			dw.name = scanner.Text()
		}
	case "file":
		if !scanner.Scan() {
			return errors.New("file takes a file path")
		}
		arg1 := scanner.Text()
		fname := filepath.Base(arg1)
		if scanner.Scan() {
			fname = scanner.Text()
		}
		// map[on_disk]on_upload
		dw.files[arg1] = fname
	case "dir":
		if !scanner.Scan() {
			return errors.New("dir takes a file path")
		}
		arg1 := scanner.Text()
		fname := filepath.Base(arg1)
		if scanner.Scan() {
			fname = scanner.Text()
		}
		// map[on_disk]on_upload
		dw.dirs[arg1] = fname
	case "deb":
		// TODO(reed):
	case "gem":
		if !scanner.Scan() {
			return errors.New("gem takes a gem")
		}
		gem := scanner.Text()
		version := ">=0"
		if scanner.Scan() {
			version = scanner.Text()
		}
		dw.gems[gem] = version
	case "gemfile":
		// TODO(reed):
	case "jar":
		// TODO(reed):
	case "pip":
		// TODO(reed):
	default:
		return errors.New(key + " not a valid .worker field")
	}
	return nil
}
