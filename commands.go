package main

// Contains each command and its configuration

// TODO(reed): fix: empty schedule payload not working ?

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
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

	bc.hud_URL_str = `Check 'https://hud.iron.io/tq/projects/` + bc.wrkr.Settings.ProjectId + "/"

	fmt.Println(LINES, `Configuring client`)
	fmt.Println(BLANKS, `Project id="`+bc.wrkr.Settings.ProjectId+`"`)
	return nil
}

type UploadCmd struct {
	command

	// TODO(reed): config flag ?
	config       *string
	maxConc      *int
	retries      *int
	retriesDelay *int
	codes        worker.Code
}

type RunCmd struct {
	command

	payload     *string
	payloadFile *string

	containerPath string
	pload         string // final payload
	codes         worker.Code
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
		return errors.New("error: schedule takes one argument")
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
	fmt.Println(LINES, "Scheduling task")

	ids, err := s.wrkr.Schedule(s.sched)
	if err != nil {
		fmt.Println(BLANKS, err)
		return
	}
	id := ids[0]

	fmt.Printf("%s scheduled %s with id: %s\n", BLANKS, s.sched.CodeName, id)
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

	err := q.flags.Parse(args)
	if err != nil {
		return err
	}

	return q.flags.validateAllFlags()
}

// Takes 1 arg for worker name
func (q *QueueCmd) Args() error {
	if q.flags.NArg() != 1 {
		return errors.New("error: queue takes one argument")
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
		Delay:    &delay,
		Timeout:  &timeout,
		Priority: *q.priority,
	}

	if payload != "" {
		q.task.Payload = payload
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
	fmt.Println(LINES, "Queueing task")

	ids, err := q.wrkr.TaskQueue(q.task)
	if err != nil {
		fmt.Println(BLANKS, err)
		return
	}
	id := ids[0]

	fmt.Printf("%s Queued %s with id: \"%s\"\n", BLANKS, q.task.CodeName, id)
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
		return errors.New("error: status takes one argument")
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
	fmt.Println(LINES, `Getting status of task with id "`+s.taskID+`"`)
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
		return errors.New("error: log takes one argument")
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
	fmt.Println(LINES, "Getting log for task with id", l.taskID)
	out, err := l.wrkr.TaskLog(l.taskID)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(out))
}

func (u *UploadCmd) Flags(args ...string) error {
	u.flags = NewWorkerFlagSet(u.Usage())
	u.maxConc = u.flags.maxConc()
	u.retries = u.flags.retries()
	u.retriesDelay = u.flags.retriesDelay()
	u.config = u.flags.config()

	err := u.flags.Parse(args)
	if err != nil {
		return err
	}
	return u.flags.validateAllFlags()
}

// takes parameter with name for worker
func (u *UploadCmd) Args() error {
	if u.flags.NArg() < 1 {
		return errors.New("error: upload takes one argument")
	}
	// TODO(reed): camel_case thing
	worker := u.flags.Arg(0) + ".worker"
	if _, err := os.Stat(worker); os.IsNotExist(err) {
		return err
	}
	// TODO(reed): turnkey
	var err error
	u.codes, err = bundleCodes(worker)
	if err != nil {
		return err
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
	return nil
}

func (u *UploadCmd) Usage() func() {
	return func() {
		fmt.Fprintln(os.Stderr, `usage: iron_worker upload [OPTIONS] worker`)
		u.flags.PrintDefaults()
	}
}

func (u *UploadCmd) Run() {
	fmt.Println(LINES, `Uploading worker "`+u.codes.Name+`"`)
	id, err := u.wrkr.CodePackageUpload(u.codes)

	if err != nil {
		fmt.Println(err)
		return
	}
	id = string(id)
	fmt.Println(BLANKS, `Code package uploaded code with id="`+id+`"`)
	fmt.Println(BLANKS, u.hud_URL_str+"code/"+id+INFO)
}

func (r *RunCmd) Flags(args ...string) error {
	r.flags = NewWorkerFlagSet(r.Usage())
	r.payload = r.flags.payload()
	r.payloadFile = r.flags.payloadFile()
	err := r.flags.Parse(args)
	if err != nil {
		return err
	}
	return r.flags.validateAllFlags()
}

func (r *RunCmd) Args() error {
	if r.flags.NArg() < 1 {
		return errors.New("error: run takes one argument")
	}

	// TODO(reed): camel_case thing
	worker := r.flags.Arg(0) + ".worker"
	if _, err := os.Stat(worker); os.IsNotExist(err) {
		return err
	}
	// TODO(reed): turnkey
	var err error
	r.codes, err = bundleCodes(worker)
	if err != nil {
		return err
	}

	payload := *r.payload
	if *r.payloadFile != "" {
		pload, err := ioutil.ReadFile(*r.payloadFile)
		if err != nil {
			return err
		}
		payload = string(pload)
	}
	r.pload = payload

	return nil
}

func (r *RunCmd) Usage() func() {
	return func() {
		fmt.Fprintln(os.Stderr, `usage: iron_worker run [OPTIONS] worker`)
		r.flags.PrintDefaults()
	}
}

// TODO(reed): config?
func (r *RunCmd) Run() {
	fmt.Println(LINES, `Running worker "`+r.codes.Name+`" locally`)

	var err error // just print errs, since there's no turning back
	r.containerPath, err = ioutil.TempDir("", "iron-worker-"+r.flags.Arg(0)+"-")
	if err != nil {
		fmt.Println(err)
	}
	defer os.RemoveAll(r.containerPath)

	err = ioutil.WriteFile(r.containerPath+"/__payload__", []byte(r.pload), 0666)
	if err != nil {
		fmt.Println(err)
	}
	for file, bytes := range r.codes.Source {
		if err := os.MkdirAll(filepath.Dir(r.containerPath+"/"+file), 0755); err != nil {
			fmt.Println(err)
		}
		if err := ioutil.WriteFile(r.containerPath+"/"+file, bytes, 0666); err != nil {
			fmt.Println(err)
		}
	}

	out, _ := exec.Command("sh", r.containerPath+"/__runner__.sh",
		"-payload", r.containerPath+"/__payload__",
		"-config", r.containerPath+"/__config__",
		"-id", "0").CombinedOutput()

	fmt.Println(string(out))

	os.RemoveAll(r.containerPath)
}
