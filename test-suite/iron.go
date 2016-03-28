package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

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

func cleanIronTaskIdAndTimestamp(output string) (string, error) {
	var buf bytes.Buffer
	var taskId string = ""
	// expecting request id as hex of bson_id
	requestIdPattern, _ := regexp.Compile("[a-f0-9]{24}")
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()

		if taskId == "" {
			parts := strings.Fields(line)

			// generic logging through logger.info, logger.warning & etc
			if len(parts) >= 3 {
				requestIdCandidate := parts[2]
				if requestIdPattern.MatchString(requestIdCandidate) {
					taskId = requestIdCandidate
				}
			}
		}

		line, isOk := util.RemoveTimestampAndRequestIdFromIronLogLine(line, taskId)
		if isOk {
			buf.WriteString(line)
			buf.WriteRune('\n')
		}
		if err := scanner.Err(); err != nil {
			return "", err
		}
	}

	return buf.String(), nil
}

func cleanIron(output []byte) ([]byte, error) {
	output = cleanIronGeneric(output)
	cleaned, err := cleanIronTaskIdAndTimestamp(string(output))
	return []byte(cleaned), err
}

//Returns a result and a debug channels. The channels are closed on test run finalization
func runOnIron(w *worker.Worker, test *util.TestDescription) (<-chan string, <-chan string) {
	result := make(chan string, 1)
	debug := make(chan string, 1)
	go func() {
		defer close(result)
		defer close(debug)
		var imagePrefix string
		if imagePrefix = os.Getenv("IRON_LAMBDA_TEST_IMAGE_PREFIX"); imagePrefix == "" {
			log.Fatalf("IRON_LAMBDA_TEST_IMAGE_PREFIX not set")
		}

		payload, _ := json.Marshal(test.Event)
		timeout := time.Duration(test.Timeout) * time.Second

		debug <- "Enqueuing the task"
		taskids, err := w.TaskQueue(worker.Task{
			Cluster:  "internal",
			CodeName: fmt.Sprintf("%s/%s", imagePrefix, test.Name),
			Payload:  string(payload),
			Timeout:  &timeout,
		})

		if err != nil {
			debug <- fmt.Sprintf("Error queueing task %s", err)
			return
		}

		if len(taskids) < 1 {
			debug <- "Something went wrong, empty taskids list"
			return
		}

		end := time.After(timeout)
		taskid := taskids[0]
		debug <- fmt.Sprintf("Task Id: %s", taskid)

		debug <- "Waiting for task"
		select {
		case <-w.WaitForTask(taskid):
		case <-end:
			debug <- fmt.Sprintf("Task timed out %s", taskid)
			return
		}

		var iron_log []byte
		debug <- "Waiting for task log"
		select {
		case _iron_log, wait_log_ok := <-w.WaitForTaskLog(taskid):
			if !wait_log_ok {
				debug <- fmt.Sprintf("Something went wrong, no task log %s", taskid)
				return
			}
			iron_log = _iron_log
		case <-end:
			debug <- fmt.Sprintf("Task timed out to get task log or the log is empty %s", taskid)
			return
		}

		cleanedLog, err := cleanIron(iron_log)
		if err != nil {
			debug <- fmt.Sprintf("Error processing a log %s", test.Name)
		} else {
			result <- string(cleanedLog)
		}
	}()
	return result, debug
}
