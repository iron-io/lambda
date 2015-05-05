package main

import (
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

func (wf *WorkerFlags) name() *string {
	return wf.String("name", "", "override code package name")
}

func (wf *WorkerFlags) payload() *string {
	return wf.String("payload", "", "give worker payload")
}

func (wf *WorkerFlags) configFile() *string {
	return wf.String("config-file", "", "upload file for worker config")
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

func (wf *WorkerFlags) retries() *int {
	return wf.Int("retries", 0, "max times to retry failed task, max 10, default 0")
}

func (wf *WorkerFlags) retriesDelay() *int {
	return wf.Int("retries-delay", 0, "time between retries, in seconds. default 0")
}

func (wf *WorkerFlags) config() *string {
	return wf.String("config", "", "provide config string (re: JSON/YAML) that will be available in file on upload")
}

func (wf *WorkerFlags) stack() *string {
	return wf.String("stack", "", "DEPRECATED: pass in image instead.")
}

func (wf *WorkerFlags) zip() *string {
	return wf.String("zip", "", "optional: name of zip file where code resides")
}

// TODO(reed): pretty sure there's a better way to get types from flags...
func (wf *WorkerFlags) validateAllFlags() error {
	if timeout := wf.Lookup("timeout"); timeout != nil {
		_, err := strconv.Atoi(timeout.Value.String())
		if err != nil {
			return err
		}
	}

	if configFile := wf.Lookup("config-file"); configFile != nil {
		configFile := configFile.Value.String()
		if configFile != "" {
			if _, err := os.Stat(configFile); os.IsNotExist(err) {
				return err
			}
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
		_, err := strconv.Atoi(priority.Value.String())
		if err != nil {
			return err
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
