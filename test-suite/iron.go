package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	"github.com/iron-io/iron_go3/worker"
	"github.com/iron-io/lambda/test-suite/util"
)

func cleanIron(input []byte) []byte {
	// We can only deal with output from downloading an image.
	if bytes.HasPrefix(input, []byte("Unable to find image")) {
		// Skip everything until line 'Status: Downloaded...'
		idx := bytes.Index(input, []byte("Status: Downloaded"))
		if idx >= 0 {
			tmp := input[idx:]
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
	return input
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
	output.Write(cleanIron(iron_log))
}
