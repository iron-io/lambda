package main

// Contains each command and its configuration

// TODO(reed): fix: empty schedule payload not working ?

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/iron-io/iron_go/config"
	"github.com/iron-io/iron_go/worker"
)

// TODO(reed): default flags for everybody
//--config CONFIG              config file

// The idea is:
//     parse flags -- if help, Usage() && quit
//  -> validate arguments, configure command
//  -> configure client
//  -> run command
//
//  if anything goes wrong, peace
type Command interface {
	Flags(...string) error // parse subcommand specific flags
	Args() error           // validate arguments
	Config() error         // configure env variables
	Usage() func()         // custom command help TODO(reed): all local now?
	Run()                  // cmd specific
}

// A command is the base for all commands implementing the Command interface.
type command struct {
	wrkr        worker.Worker
	flags       *WorkerFlags
	hud_URL_str string
	token       *string
	projectID   *string
}

// All Commands will do similar configuration
func (bc *command) Config() error {
	bc.wrkr.Settings = config.ConfigWithEnv("iron_worker", *envFlag)
	if *projectIDFlag != "" {
		bc.wrkr.Settings.ProjectId = *projectIDFlag
	}
	if *tokenFlag != "" {
		bc.wrkr.Settings.Token = *tokenFlag
	}

	if bc.wrkr.Settings.ProjectId == "" {
		return errors.New("did not find project id in any config files or env variables")
	}
	if bc.wrkr.Settings.Token == "" {
		return errors.New("did not find token in any config files or env variables")
	}

	bc.hud_URL_str = `Check https://hud.iron.io/tq/projects/` + bc.wrkr.Settings.ProjectId + "/"

	fmt.Println(LINES, `Configuring client`)

	pName, err := projectName(bc.wrkr.Settings)
	if err != nil {
		return err
	}

	fmt.Printf(`%s Project '%s' with id='%s'`, BLANKS, pName, bc.wrkr.Settings.ProjectId)
	fmt.Println()
	return nil
}

func projectName(config config.Settings) (string, error) {
	// get project name -- go api won't play ball
	resp, err := http.Get(fmt.Sprintf("%s://%s:%d/%s/projects/%s?oauth=%s",
		config.Scheme, config.Host, config.Port,
		config.ApiVersion, config.ProjectId, config.Token))

	if err != nil {
		return "", err
	}

	var reply struct {
		Name string `json:"name"`
	}
	err = json.NewDecoder(resp.Body).Decode(&reply)
	return reply.Name, err
}

type UploadCmd struct {
	command

	name         *string
	config       *string
	configFile   *string
	stack        *string // deprecated
	maxConc      *int
	retries      *int
	retriesDelay *int
	zip          *string
	codes        worker.Code // for fields, not code
	cmd          string
}

type QueueCmd struct {
	command

	// flags
	payload     *string
	payloadFile *string
	priority    *int
	timeout     *int
	delay       *int
	wait        *bool
	cluster     *string

	// payload
	task worker.Task
}

type SchedCmd struct {
	command
	payload     *string
	payloadFile *string
	priority    *int
	timeout     *int
	delay       *int
	maxConc     *int
	runEvery    *int
	runTimes    *int
	endAt       *string // time.RubyTime
	startAt     *string // time.RubyTime

	sched worker.Schedule
}

type StatusCmd struct {
	command
	taskID string
}

type LogCmd struct {
	command
	taskID string
}

func (s *SchedCmd) Flags(args ...string) error {
	s.flags = NewWorkerFlagSet(s.Usage())

	s.payload = s.flags.payload()
	s.payloadFile = s.flags.payloadFile()
	s.priority = s.flags.priority()
	s.timeout = s.flags.timeout()
	s.delay = s.flags.delay()
	s.maxConc = s.flags.maxConc()
	s.runEvery = s.flags.runEvery()
	s.runTimes = s.flags.runTimes()
	s.endAt = s.flags.endAt()
	s.startAt = s.flags.startAt()

	err := s.flags.Parse(args)
	if err != nil {
		return err
	}

	return s.flags.validateAllFlags()
}

func (s *SchedCmd) Args() error {
	if s.flags.NArg() != 1 {
		return errors.New("error: schedule takes one argument, a code name")
	}

	delay := time.Duration(*s.delay) * time.Second

	s.sched = worker.Schedule{
		CodeName: s.flags.Arg(0),
		Delay:    &delay,
		Priority: s.priority,
		RunTimes: s.runTimes,
	}

	payload := *s.payload
	if *s.payloadFile != "" {
		pload, err := ioutil.ReadFile(*s.payloadFile)
		if err != nil {
			return err
		}
		payload = string(pload)
	}

	if payload != "" {
		s.sched.Payload = payload
	}

	if *s.endAt != "" {
		t, _ := time.Parse(time.RubyDate, *s.endAt) // checked in validateFlags()
		s.sched.EndAt = &t
	}
	if *s.startAt != "" {
		t, _ := time.Parse(time.RubyDate, *s.startAt)
		s.sched.StartAt = &t
	}
	if *s.maxConc > 0 {
		s.sched.MaxConcurrency = s.maxConc
	}
	if *s.runEvery > 0 {
		s.sched.RunEvery = s.runEvery
	}

	return nil
}

func (s *SchedCmd) Usage() func() {
	return func() {
		fmt.Fprintln(os.Stderr, `usage: iron_worker schedule [OPTIONS] CODE_PACKAGE_NAME`)
		s.flags.PrintDefaults()
	}
}

func (s *SchedCmd) Run() {
	fmt.Println(LINES, "Scheduling task '"+s.sched.CodeName+"'")

	ids, err := s.wrkr.Schedule(s.sched)
	if err != nil {
		fmt.Println(BLANKS, err)
		return
	}
	id := ids[0]

	fmt.Printf("%s Scheduled task with id='%s'\n", BLANKS, id)
	fmt.Println(BLANKS, s.hud_URL_str+"scheduled_jobs/"+id+INFO)
}

func (q *QueueCmd) Flags(args ...string) error {
	q.flags = NewWorkerFlagSet(q.Usage())

	q.payload = q.flags.payload()
	q.payloadFile = q.flags.payloadFile()
	q.priority = q.flags.priority()
	q.timeout = q.flags.timeout()
	q.delay = q.flags.delay()
	q.wait = q.flags.wait()
	q.cluster = q.flags.cluster()

	err := q.flags.Parse(args)
	if err != nil {
		return err
	}

	return q.flags.validateAllFlags()
}

// Takes 1 arg for worker name
func (q *QueueCmd) Args() error {
	if q.flags.NArg() != 1 {
		return errors.New("error: queue takes one argument, a code name")
	}

	payload := *q.payload
	if *q.payloadFile != "" {
		pload, err := ioutil.ReadFile(*q.payloadFile)
		if err != nil {
			return err
		}
		payload = string(pload)
	}

	delay := time.Duration(*q.delay) * time.Second
	timeout := time.Duration(*q.timeout) * time.Second

	q.task = worker.Task{
		CodeName: q.flags.Arg(0),
		Payload:  payload,
		Priority: *q.priority,
		Timeout:  &timeout,
		Delay:    &delay,
		Cluster:  *q.cluster,
	}

	return nil
}

func (q *QueueCmd) Usage() func() {
	return func() {
		fmt.Fprintln(os.Stderr, `usage: iron_worker queue [OPTIONS] CODE_PACKAGE_NAME`)
		q.flags.PrintDefaults()
	}
}

func (q *QueueCmd) Run() {
	fmt.Println(LINES, "Queueing task '"+q.task.CodeName+"'")

	ids, err := q.wrkr.TaskQueue(q.task)
	if err != nil {
		fmt.Println(BLANKS, err)
		return
	}
	id := ids[0]

	fmt.Printf("%s Queued task with id='%s'\n", BLANKS, id)
	fmt.Println(BLANKS, q.hud_URL_str+"jobs/"+id+INFO)

	if *q.wait {
		fmt.Println(LINES, "Waiting for task", id)

		out := q.wrkr.WaitForTaskLog(id)

		log := <-out
		fmt.Println(LINES, "Done")
		fmt.Println(LINES, "Printing Log:")
		fmt.Printf("%s", string(log))
	}
}

func (s *StatusCmd) Flags(args ...string) error {
	s.flags = NewWorkerFlagSet(s.Usage())
	err := s.flags.Parse(args)
	if err != nil {
		return err
	}

	return s.flags.validateAllFlags()
}

// Takes one parameter, the task_id to acquire status of
func (s *StatusCmd) Args() error {
	if s.flags.NArg() != 1 {
		return errors.New("error: status takes one argument, a task_id")
	}
	s.taskID = s.flags.Arg(0)
	return nil
}

func (s *StatusCmd) Usage() func() {
	return func() {
		fmt.Fprintln(os.Stderr, `usage: iron_worker status [OPTIONS] task_id`)
		s.flags.PrintDefaults()
	}
}

func (s *StatusCmd) Run() {
	fmt.Println(LINES, `Getting status of task with id='`+s.taskID+`'`)
	taskInfo, err := s.wrkr.TaskInfo(s.taskID)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(BLANKS, taskInfo.Status)
}

func (l *LogCmd) Flags(args ...string) error {
	l.flags = NewWorkerFlagSet(l.Usage())
	err := l.flags.Parse(args)
	if err != nil {
		return err
	}
	return l.flags.validateAllFlags()
}

// Takes one parameter, the task_id to log
func (l *LogCmd) Args() error {
	if l.flags.NArg() < 1 {
		return errors.New("error: log takes one argument, a task_id")
	}
	l.taskID = l.flags.Arg(0)
	return nil
}

func (l *LogCmd) Usage() func() {
	return func() {
		fmt.Fprintln(os.Stderr, `usage: iron_worker log [OPTIONS] task_id`)
		l.flags.PrintDefaults()
	}
}

func (l *LogCmd) Run() {
	fmt.Println(LINES, "Getting log for task with id='"+l.taskID+"'")
	out, err := l.wrkr.TaskLog(l.taskID)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(out))
}

func (u *UploadCmd) Flags(args ...string) error {
	u.flags = NewWorkerFlagSet(u.Usage())
	u.name = u.flags.name()
	u.stack = u.flags.stack()
	u.maxConc = u.flags.maxConc()
	u.retries = u.flags.retries()
	u.retriesDelay = u.flags.retriesDelay()
	u.config = u.flags.config()
	u.configFile = u.flags.configFile()
	u.zip = u.flags.zip()

	err := u.flags.Parse(args)
	if err != nil {
		return err
	}
	return u.flags.validateAllFlags()
}

// `iron worker upload [--zip [ZIPFILE]] [IMAGE] [[COMMAND]...]`
//
// old deprecated: `iron worker upload [ZIPFILE] [COMMAND]`
func (u *UploadCmd) Args() error {
	if u.flags.NArg() < 2 {
		return errors.New("upload takes at least two arguments, the name of the worker and the name of the Docker image. eg: iron worker upload [--zip WORKER_ZIP] WORKER_NAME DOCKER_IMAGE [COMMAND]")
	}

	u.codes.Command = strings.TrimSpace(strings.Join(u.flags.Args()[2:], " "))
	if *u.stack != "" {
		// deprecated
		u.codes.Stack = *u.stack
		*u.zip = u.flags.Arg(0)
		u.codes.Command = strings.TrimSpace(strings.Join(u.flags.Args()[1:], " "))
	} else {
		u.codes.Image = u.flags.Arg(1)
		u.codes.Command = strings.TrimSpace(strings.Join(u.flags.Args()[2:], " "))
		// command also optional, filled in above
		// zip filled in from flag, optional
	}

	if *u.name == "" {
		if *u.stack != "" {
			u.codes.Name = strings.TrimSuffix(filepath.Base(*u.zip), ".zip")
		} else {
			u.codes.Name = u.flags.Arg(0)
		}
	} else {
		u.codes.Name = *u.name
	}

	if *u.zip != "" {
		if u.codes.Command == "" { // must have command if using zip
			return errors.New("uploading a zip file requires a command, see -help")
		}
		// make sure it exists and it's a zip
		if !strings.HasSuffix(*u.zip, ".zip") {
			return errors.New("file extension must be .zip, got: " + *u.zip)
		}
		if _, err := os.Stat(*u.zip); err != nil {
			return err
		}
	}
	if *u.maxConc > 0 {
		u.codes.MaxConcurrency = *u.maxConc
	}
	if *u.retries > 0 {
		u.codes.Retries = *u.retries
	}
	if *u.retriesDelay > 0 {
		u.codes.RetriesDelay = time.Duration(*u.retriesDelay) * time.Second
	}
	if *u.config != "" {
		u.codes.Config = *u.config
	}

	if *u.configFile != "" {
		pload, err := ioutil.ReadFile(*u.configFile)
		if err != nil {
			return err
		}
		u.codes.Config = string(pload)
	}

	return nil
}

func (u *UploadCmd) Usage() func() {
	return func() {
		fmt.Fprintln(os.Stderr, `usage: iron_worker upload -name myworker [OPTIONS] worker.zip command...`)
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, `or`)
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, `usage: iron_worker upload -name myworker [-zip my.zip] [OPTIONS] some/image [command...]`)
		u.flags.PrintDefaults()
	}
}

func (u *UploadCmd) Run() {
	fmt.Println(LINES, `Uploading worker '`+u.codes.Name+`'`)
	id, err := pushCodes(*u.zip, &u.wrkr, u.codes)

	if err != nil {
		fmt.Println(err)
		return
	}
	id = string(id)
	fmt.Println(BLANKS, `Uploaded code package with id='`+id+`'`)
	fmt.Println(BLANKS, u.hud_URL_str+"code/"+id+INFO)
}
