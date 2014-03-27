// CodePackageUpload uploads a code package
package main

import (
	"bufio"
	"errors"
	"io/ioutil"
	"os"
	"strings"

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

func (dw *dotWorker) code() (worker.Code, error) {
	codes := worker.Code{
		Name:     dw.name,
		Runtime:  `sh`,
		FileName: `__runner__.sh`,
	}

	exec, err := ioutil.ReadFile(dw.exec)
	if err != nil {
		return codes, err
	}

	runner := []byte(RUNNER +
		runtime(dw.runtime, dw.exec) + ` \"$@\"`) //+`#{File.basename(@exec.path)} #{params}

	// TODO(reed): pass or receive one of these from runtime() somehow
	source := worker.CodeSource{
		dw.exec:         exec,
		`__runner__.sh`: runner,
	}

	//for f, storeAs := range dw.files {
	//}
	//for dir, storeAs := range dw.dirs {

	//}

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
	dw := &dotWorker{
		name: dotWorkerFile[:len(dotWorkerFile)-7], // TODO(reed): camel_case ?
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
		words := strings.Fields(line)
		err := dw.parseLine(words)
		if err != nil {
			return nil, err
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return dw, nil
}

func rmQuotes(s string) string {
	return s[1 : len(s)-1]
}

// TODO(reed): don't tokenize lines at a time, build real lexer
func (dw *dotWorker) parseLine(line []string) error {
	switch line[0] {
	case "runtime":
		if len(line) < 2 {
			return errors.New("runtime takes one arg")
		}
		runtime := rmQuotes(line[1])
		switch runtime {
		// TODO(reed): ehhh, duplication of concerns? server take care of this?
		case "binary", "go", "java", "mono", "node", "php", "python", "ruby":
			dw.runtime = runtime
		default:
			return errors.New(runtime + " not a valid runtime")
		}
	case "stack":
		if len(line) < 2 {
			return errors.New("stack takes one arg")
		}
		dw.stack = rmQuotes(line[1])
	case "name":
	case "set_env":
	case "full_remote_build":
	case "build":
	case "exec":
		if len(line) < 2 {
			return errors.New("exec takes a file path")
		}
		dw.exec = rmQuotes(line[1])
		if len(line) == 3 {
			dw.name = rmQuotes(line[2])
		}
	case "file":
		if err := dw.parseFile(line); err != nil {
			return err
		}
	case "dir":
		if err := dw.parseDir(line); err != nil {
			return err
		}
	case "deb":
	case "gem":
	case "gemfile":
	case "jar":
	case "pip":
	default:
		return errors.New(line[0] + " not a valid .worker field")
	}
	return nil
}

func (dw *dotWorker) parseFile(line []string) error {
	if len(line) < 2 {
		return errors.New("file requires path")
	}
	fname := line[1]
	if len(line) == 3 {
		fname = line[2]
	}
	dw.files[line[1]] = fname
	return nil
}

func (dw *dotWorker) parseDir(line []string) error {
	if len(line) < 2 {
		return errors.New("dir requires path")
	}
	dirName := line[1]
	if len(line) == 3 {
		dirName = line[2]
	}
	dw.dirs[line[1]] = dirName
	return nil
}
