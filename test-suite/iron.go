package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/iron-io/iron_go3/worker"
	"github.com/iron-io/lambda/test-suite/util"
)

func cleanIronGeneric(output []byte) []byte {
	// We can only deal with output from downloading an image.
	if bytes.HasPrefix(output, []byte("Unable to find image")) {
		// Skip everything until line 'Status: Downloaded...'
		idx := bytes.Index(output, []byte("Status: Downloaded"))
		if idx >= 0 {
			tmp := output[idx:]
			lineidx := bytes.IndexByte(tmp, '\n')
			if lineidx >= 0 {
				// Skip the newline itself.
				return tmp[lineidx+1:]
			} else {
				// Could not find newline, so this was the last line.
				return []byte{}
			}
		}
	}
	return output
}

func cleanPython27IronOutput(output string) (string, error) {
	var buf bytes.Buffer
	var requestId string = ""
	requestIdPattern, _ := regexp.Compile("[a-f0-9]{24}")
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()

		if requestId == "" {
			parts := strings.Fields(line)

			// generic logging through logger.info, logger.warning & etc
			if len(parts) >= 3 {
				requestIdCandidate := parts[2]
				if requestIdPattern.MatchString(requestIdCandidate) {
					requestId = requestIdCandidate
				}
			}
		}

		line = util.RemoveTimestampAndRequestIdFromLogLine(line, requestId)

		buf.WriteString(line)
		buf.WriteRune('\n')
		if err := scanner.Err(); err != nil {
			return "", err
		}
	}

	return buf.String(), nil
}

func convertUtf8ByteSequenceToString(buffer []byte) (string, error) {
	if !utf8.Valid(buffer) {
		return "", errors.New("Invalid utf8 sequence")
	}

	result := make([]rune, 0, utf8.RuneCount(buffer))

	for len(buffer) > 0 {
		r, c := utf8.DecodeRune(buffer)

		result = append(result, r)
		buffer = buffer[c:]
	}
	return string(result), nil
}

func convertStringToUtf8BySequence(buffer string) []byte {
	var buf bytes.Buffer
	runeBuffer := make([]byte, 32)
	for _, r := range buffer {
		l := utf8.EncodeRune(runeBuffer, r)
		buf.Write(runeBuffer[:l])
	}

	return buf.Bytes()
}

func cleanIron(runtime string, output []byte) ([]byte, error) {
	output = cleanIronGeneric(output)
	switch runtime {
	case "python2.7":
		{
			outputAsString, err := convertUtf8ByteSequenceToString(output)
			if err != nil {
				return nil, err
			}
			cleaned, err := cleanPython27IronOutput(outputAsString)
			if err != nil {
				return nil, err
			}
			return convertStringToUtf8BySequence(cleaned), nil
		}
	default:
		return output, nil
	}
}

func runOnIron(w *worker.Worker, wg *sync.WaitGroup, test *util.TestDescription, result chan<- io.Reader) {
	var imagePrefix string
	if imagePrefix = os.Getenv("IRON_LAMBDA_TEST_IMAGE_PREFIX"); imagePrefix == "" {
		log.Fatalf("IRON_LAMBDA_TEST_IMAGE_PREFIX not set")
	}

	var output bytes.Buffer
	defer func() {
		result <- &output
		wg.Done()
	}()

	payload, _ := json.Marshal(test.Event)

	taskids, err := w.TaskQueue(worker.Task{
		CodeName: fmt.Sprintf("%s/%s", imagePrefix, test.Name),
		Payload:  string(payload),
	})

	if err != nil {
		output.WriteString(fmt.Sprintf("Error queueing task %s %s", test.Name, err))
		return
	}

	if len(taskids) < 1 {
		output.WriteString(fmt.Sprintf("Something went wrong, empty taskids list", test.Name))
		return
	}

	taskid := taskids[0]

	<-w.WaitForTask(taskid)
	iron_log := <-w.WaitForTaskLog(taskid)
	cleanedLog, err := cleanIron(test.Runtime, iron_log)
	if err != nil {
		output.WriteString(fmt.Sprintf("Error processing a log for task %s %s", test.Name, err))
	} else {
		output.Write(cleanedLog)
	}
}
