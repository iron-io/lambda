package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type WorkerFlags struct {
	*flag.FlagSet
}

func NewWorkerFlagSet() *WorkerFlags {
	flags := flag.NewFlagSet("command", flag.ContinueOnError)
	flags.Usage = func() {}
	return &WorkerFlags{flags}
}

func (wf *WorkerFlags) name() *string {
	return wf.String("name", "", "override code package name")
}
func (wf *WorkerFlags) dockerRepoPass() *string {
	return wf.String("p", "", "docker repo password")
}
func (wf *WorkerFlags) dockerRepoUserName() *string {
	return wf.String("u", "", "docker repo user name")
}
func (wf *WorkerFlags) dockerRepoUrl() *string {
	return wf.String("url", "", "docker repo url, if you're using custom repo")
}
func (wf *WorkerFlags) dockerRepoEmail() *string {
	return wf.String("e", "", "docker repo user email")
}

func (wf *WorkerFlags) host() *string {
	return wf.String("host", "", "paas host")
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
	return wf.Int("priority", -3, "0(default), 1 or 2; uses worker's default priority if unset")
}

func (wf *WorkerFlags) defaultPriority() *int {
	return wf.Int("default-priority", -3, "0(default), 1 or 2")
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
	return wf.Int("run-times", 0, "number of times a task will run")
}

func (wf *WorkerFlags) endAt() *string {
	return wf.String("end-at", "", "time or datetime in RFC3339 format: '2006-01-02T15:04:05Z07:00'")
}

func (wf *WorkerFlags) startAt() *string {
	return wf.String("start-at", "", "time or datetime in RFC3339 format: '2006-01-02T15:04:05Z07:00'")
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

func (wf *WorkerFlags) zip() *string {
	return wf.String("zip", "", "optional: name of zip file where code resides")
}

func (wf *WorkerFlags) cluster() *string {
	return wf.String("cluster", "", "optional: specify cluster to queue task on")
}

func (wf *WorkerFlags) label() *string {
	return wf.String("label", "", "optional: specify label for a task")
}

func (wf *WorkerFlags) encryptionKey() *string {
	return wf.String("encryption-key", "", "optional: specify a hex encoded encryption key")
}

// -- envSlice Value
type envVariable struct {
	Name  string
	Value string
}

type envSlice []envVariable

func (s *envSlice) Set(val string) error {
	if !strings.Contains(val, "=") {
		return errors.New("Environment variable format is 'ENVNAME=value'")
	}
	pair := strings.SplitN(val, "=", 2)
	envVar := envVariable{Name: pair[0], Value: pair[1]}
	*s = append(*s, envVar)
	return nil
}

func (s *envSlice) Get() interface{} {
	return *s
}

func (s *envSlice) String() string { return fmt.Sprintf("%v", *s) }

func (wf *WorkerFlags) envVars() *envSlice {
	var sameNamedFlags envSlice
	wf.Var(&sameNamedFlags, "e", "optional: specify environment variable for your code in format 'ENVNAME=value'")
	return &sameNamedFlags
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
			_, err := time.Parse(time.RFC3339, endat)
			if err != nil {
				return err
			}
		}
	}

	if startat := wf.Lookup("start-at"); startat != nil {
		startat := startat.Value.String()
		if startat != "" {
			_, err := time.Parse(time.RFC3339, startat)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

type MqFlags struct {
	*flag.FlagSet
}

func NewMqFlagSet() *MqFlags {
	flags := flag.NewFlagSet("commands", flag.ContinueOnError)
	flags.Usage = func() {}
	return &MqFlags{flags}
}

func (mf *MqFlags) validateAllFlags() error {
	if payloadFile := mf.Lookup("f"); payloadFile != nil {
		payloadFile := payloadFile.Value.String()
		if payloadFile != "" {
			if _, err := os.Stat(payloadFile); os.IsNotExist(err) {
				return err
			}
		}
	}
	return nil
}

func (mf *MqFlags) filename() *string {
	return mf.String("f", "", "optional: provide a json file of messages to be posted")
}

func (mf *MqFlags) outputfile() *string {
	return mf.String("o", "", "optional: write json output to a file")
}
func (mf *MqFlags) perPage() *int {
	return mf.Int("perPage", 30, "optional: amount of queues shown per page (default: 30)")
}
func (mf *MqFlags) page() *string {
	return mf.String("page", "0", "optional: starting page (default: 0)")
}

func (mf *MqFlags) filter() *string {
	return mf.String("filter", "", "optional: prefix filter (default: \"\")")
}

func (mf *MqFlags) n() *int {
	return mf.Int("n", 1, "optional: number of messages to get")
}

func (mf *MqFlags) timeout() *int {
	return mf.Int("t", 60, "optional: timeout until message is put back on queue")
}
func (mf *MqFlags) subscriberList() *bool {
	return mf.Bool("subscriber-list", false, "optional: printout all subscriber names and URLs")
}
