package main

import (
	"errors"
	"flag"
	"os"
	"strconv"
	"time"
)

type WorkerFlags struct {
	*flag.FlagSet
}

// TODO(reed): -help here?
func NewWorkerFlagSet(usage func()) *WorkerFlags {
	flags := flag.NewFlagSet("command", flag.ContinueOnError)
	flags.Usage = usage
	return &WorkerFlags{flags}
}

func (wf *WorkerFlags) payload() *string {
	return wf.String("payload", "", "give worker payload")
}

func (wf *WorkerFlags) payloadFile() *string {
	return wf.String("payload-file", "", "give worker payload of file contents")
}

func (wf *WorkerFlags) priority() *int {
	return wf.Int("priority", 0, "0(default), 1 or 2")
}

func (wf *WorkerFlags) timeout() *int {
	return wf.Int("timeout", 3600, "0-3600(default) max runtime for task in seconds")
}

func (wf *WorkerFlags) delay() *int {
	return wf.Int("delay", 0, "seconds to delay before queueing task")
}

func (wf *WorkerFlags) wait() *bool {
	return wf.Bool("wait", false, "wait for task to complete and print log")
}

func (wf *WorkerFlags) maxConc() *int {
	return wf.Int("max-concurrency", -1, "max workers to run in parallel. default is no limit")
}

func (wf *WorkerFlags) runEvery() *int {
	return wf.Int("run-every", -1, "time between runs in seconds (>= 60), default is run once")
}

func (wf *WorkerFlags) runTimes() *int {
	return wf.Int("run-times", 1, "number of times a task will run")
}

func (wf *WorkerFlags) endAt() *string {
	return wf.String("end-at", "", "time or datetime of form 'Mon Jan 2 15:04:05 -0700 2006'")
}

func (wf *WorkerFlags) startAt() *string {
	return wf.String("start-at", "", "time or datetime of form 'Mon Jan 2 15:04:05 -0700 2006'")
}

// TODO(reed): pretty sure there's a better way to get types from flags...
func (wf *WorkerFlags) validateAllFlags() error {
	if timeout := wf.Lookup("timeout"); timeout != nil {
		timeout, err := strconv.Atoi(timeout.Value.String())
		if err != nil {
			return err
		}
		if timeout < 0 || timeout > 3600 {
			return errors.New("timeout can only be 0-3600(default)")
		}
	}

	if payloadFile := wf.Lookup("payload-file"); payloadFile != nil {
		payloadFile := payloadFile.Value.String()
		if payloadFile != "" {
			if _, err := os.Stat(payloadFile); os.IsNotExist(err) {
				return err
			}
		}
	}

	if priority := wf.Lookup("priority"); priority != nil {
		priority, err := strconv.Atoi(priority.Value.String())
		if err != nil {
			return err
		}
		if priority < 0 || priority > 2 {
			return errors.New("priority can only be 0(default), 1, or 2")
		}
	}

	if endat := wf.Lookup("end-at"); endat != nil {
		endat := endat.Value.String()
		if endat != "" {
			_, err := time.Parse(time.RubyDate, endat)
			if err != nil {
				return err
			}
		}
	}

	if startat := wf.Lookup("start-at"); startat != nil {
		startat := startat.Value.String()
		if startat != "" {
			_, err := time.Parse(time.RubyDate, startat)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
