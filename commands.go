package main

// Contains each command and its configuration

// TODO(reed): fix: empty schedule payload not working ?

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/iron-io/iron_go3/config"
	"github.com/iron-io/iron_go3/worker"
)

// TODO(reed): default flags for everybody

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
	Usage()                // custom command help TODO(reed): all local now?
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

type DockerLoginCmd struct {
	command
	Email         *string `json:"email"`
	Username      *string `json:"username"`
	Password      *string `json:"password"`
	Serveraddress *string `json:"serveraddress"`
	TestAuth      *string `json:"-"`    //
	RemoteAuth    *string `json:"-"`    //
	Auth          string  `json:"auth"` //Docker api require to use empty auth
}

type UploadCmd struct {
	command

	name            *string
	config          *string
	configFile      *string
	maxConc         *int
	retries         *int
	retriesDelay    *int
	defaultPriority *int
	host            *string
	zip             *string
	codes           worker.Code // for fields, not code
	cmd             string
	envVars         *envSlice
}

type RegisterCmd struct {
	command

	name            *string
	config          *string
	configFile      *string
	maxConc         *int
	retries         *int
	retriesDelay    *int
	defaultPriority *int
	host            *string
	codes           worker.Code // for fields, not code
	cmd             string
	envVars         *envSlice
}

type QueueCmd struct {
	command

	// flags
	payload           *string
	payloadFile       *string
	priority          *int
	timeout           *int
	delay             *int
	wait              *bool
	cluster           *string
	label             *string
	encryptionKey     *string
	encryptionKeyFile *string

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
	cluster     *string
	endAt       *string // time.RubyTime
	startAt     *string // time.RubyTime
	label       *string

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
	s.flags = NewWorkerFlagSet()

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
	s.cluster = s.flags.cluster()
	s.label = s.flags.label()

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

	var priority *int
	if *s.priority > -3 && *s.priority < 3 {
		priority = s.priority
	}

	s.sched = worker.Schedule{
		CodeName: s.flags.Arg(0),
		Delay:    &delay,
		Priority: priority,
		RunTimes: s.runTimes,
		Cluster:  *s.cluster,
		Label:    *s.label,
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
	} else {
		s.sched.Payload = "{}" // if we don't set this, it gets a 400 from API.
	}

	if *s.endAt != "" {
		t, _ := time.Parse(time.RFC3339, *s.endAt) // checked in validateFlags()
		s.sched.EndAt = &t
	}
	if *s.startAt != "" {
		t, _ := time.Parse(time.RFC3339, *s.startAt)
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

func (s *SchedCmd) Usage() {
	fmt.Fprintln(os.Stderr, `usage: iron worker schedule [OPTIONS] CODE_PACKAGE_NAME`)
	s.flags.PrintDefaults()
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
	q.flags = NewWorkerFlagSet()

	q.payload = q.flags.payload()
	q.payloadFile = q.flags.payloadFile()
	q.priority = q.flags.priority()
	q.timeout = q.flags.timeout()
	q.delay = q.flags.delay()
	q.wait = q.flags.wait()
	q.cluster = q.flags.cluster()
	q.label = q.flags.label()
	q.encryptionKey = q.flags.encryptionKey()
	q.encryptionKeyFile = q.flags.encryptionKeyFile()

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

	var priority *int
	if *q.priority > -3 && *q.priority < 3 {
		priority = q.priority
	}

	encryptionKey := []byte(*q.encryptionKey)
	if *q.encryptionKeyFile != "" {
		var err error
		encryptionKey, err = ioutil.ReadFile(*q.encryptionKeyFile)
		if err != nil {
			return err
		}
	}
	if len(encryptionKey) > 0 {
		var err error
		payload, err = rsaEncrypt(encryptionKey, payload)
		if err != nil {
			return err
		}
	}

	q.task = worker.Task{
		CodeName: q.flags.Arg(0),
		Payload:  payload,
		Priority: priority,
		Timeout:  &timeout,
		Delay:    &delay,
		Cluster:  *q.cluster,
		Label:    *q.label,
	}

	return nil
}

func (q *QueueCmd) Usage() {
	fmt.Fprintln(os.Stderr, `usage: iron worker queue [OPTIONS] CODE_PACKAGE_NAME`)
	q.flags.PrintDefaults()
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
		fmt.Println(LINES, yellow("Waiting for task", id))

		ti := <-q.wrkr.WaitForTask(id)
		if ti.Msg != "" {
			fmt.Fprintln(os.Stderr, "error running task:", ti.Msg)
			return
		}

		log, err := q.wrkr.TaskLog(id)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error getting log:", err)
			return
		}

		fmt.Println(LINES, green("Done"))
		fmt.Println(LINES, "Printing Log:")
		fmt.Printf("%s", string(log))
	}
}

func (s *StatusCmd) Flags(args ...string) error {
	s.flags = NewWorkerFlagSet()
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

func (s *StatusCmd) Usage() {
	fmt.Fprintln(os.Stderr, `usage: iron worker status [OPTIONS] task_id`)
	s.flags.PrintDefaults()
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
	l.flags = NewWorkerFlagSet()
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

func (l *LogCmd) Usage() {
	fmt.Fprintln(os.Stderr, `usage: iron worker log [OPTIONS] task_id`)
	l.flags.PrintDefaults()
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
func (l *DockerLoginCmd) Flags(args ...string) error {
	l.flags = NewWorkerFlagSet()

	l.Email = l.flags.dockerRepoEmail()
	l.Password = l.flags.dockerRepoPass()
	l.Serveraddress = l.flags.dockerRepoUrl()
	l.Username = l.flags.dockerRepoUserName()

	err := l.flags.Parse(args)
	if err != nil {
		return err
	}
	return l.flags.validateAllFlags()
}

// Takes one parameter, the task_id to log
func (l *DockerLoginCmd) Args() error {

	if *l.Email == "" || *l.Username == "" || *l.Password == "" || l.Email == nil || l.Username == nil || l.Password == nil {
		return errors.New("you should set email(-e), password(-p), username(-u) parameters")
	}

	if *l.Serveraddress == "" || l.Serveraddress == nil {
		defaultUrl := "https://index.docker.io/v1/" //default dockerhub url
		l.Serveraddress = &defaultUrl
	}

	auth := base64.StdEncoding.EncodeToString([]byte(*l.Username + ":" + *l.Password))
	l.TestAuth = &auth

	//{"username": "string", "password": "string", "email": "string", "serveraddress" : "string", "auth": ""}
	bytes, _ := json.Marshal(*l)
	authString := base64.StdEncoding.EncodeToString(bytes)
	l.RemoteAuth = &authString

	req, err := http.NewRequest("GET", *l.Serveraddress+"users/", nil)
	if err != nil {
		return fmt.Errorf("error authenticating docker login: %v", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Encoding", "gzip/deflate")
	req.Header.Set("Authorization", "Basic "+*l.TestAuth)
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)

	if err != nil || res.StatusCode != 200 {
		return errors.New("Docker repo auth failed")
	}

	return nil
}

func (l *DockerLoginCmd) Usage() {
	fmt.Fprintln(os.Stderr, `usage: iron docker login -u -p -e -url`)
	l.flags.PrintDefaults()
}

func (l *DockerLoginCmd) Run() {
	fmt.Println(LINES, "Storing docker repo credentials")
	auth := map[string]string{
		"auth": *l.RemoteAuth,
	}
	msg, err := dockerLogin(&l.wrkr, &auth)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(BLANKS, green(`Added docker repo credentials: `+msg))
}

func (u *UploadCmd) Flags(args ...string) error {
	u.flags = NewWorkerFlagSet()
	u.name = u.flags.name()
	u.maxConc = u.flags.maxConc()
	u.retries = u.flags.retries()
	u.retriesDelay = u.flags.retriesDelay()
	u.defaultPriority = u.flags.defaultPriority()
	u.config = u.flags.config()
	u.configFile = u.flags.configFile()
	u.zip = u.flags.zip()
	u.envVars = u.flags.envVars()

	err := u.flags.Parse(args)
	if err != nil {
		return err
	}
	return u.flags.validateAllFlags()
}

// `iron worker upload [--zip ZIPFILE] --name NAME IMAGE [COMMAND]`
func (u *UploadCmd) Args() error {
	if u.flags.NArg() < 1 {
		return errors.New("command takes at least one argument. see -help")
	}

	u.codes.Command = strings.TrimSpace(strings.Join(u.flags.Args()[1:], " "))
	u.codes.Image = u.flags.Arg(0)

	if *u.name == "" {
		return errors.New("must specify -name for your worker")
	} else {
		u.codes.Name = *u.name
	}

	if *u.zip != "" {
		// make sure it exists and it's a zip
		if !strings.HasSuffix(*u.zip, ".zip") {
			return errors.New("file extension must be .zip, got: " + *u.zip)
		}
		if _, err := os.Stat(*u.zip); err != nil {
			return err
		}
	}
	u.codes.MaxConcurrency = *u.maxConc
	u.codes.Retries = *u.retries
	u.codes.RetriesDelay = *u.retriesDelay
	u.codes.Config = *u.config
	u.codes.DefaultPriority = *u.defaultPriority

	if u.host != nil && *u.host != "" {
		u.codes.Host = *u.host
	}

	if *u.configFile != "" {
		pload, err := ioutil.ReadFile(*u.configFile)
		if err != nil {
			return err
		}
		u.codes.Config = string(pload)
	}

	if *u.envVars != nil {
		if envSlice, ok := u.envVars.Get().(envSlice); ok {
			envVarsMap := make(map[string]string, len(envSlice))
			for _, envItem := range envSlice {
				envVarsMap[envItem.Name] = envItem.Value
			}
			u.codes.EnvVars = envVarsMap
		}
	}

	return nil
}

func (u *UploadCmd) Usage() {
	fmt.Fprintln(os.Stderr, `usage: iron worker upload [-zip my.zip] -name NAME [OPTIONS] some/image[:tag] [command...]`)
	u.flags.PrintDefaults()
}

func (u *UploadCmd) Run() {
	if u.codes.Host != "" {
		fmt.Println(LINES, `Spinning up '`+u.codes.Name+`'`)
	} else {
		fmt.Println(LINES, `Uploading worker '`+u.codes.Name+`'`)
	}
	code, err := pushCodes(*u.zip, &u.wrkr, u.codes)
	if err != nil {
		fmt.Println(err)
		return
	}
	if code.Host != "" {
		fmt.Println(BLANKS, green(`Hosted at: '`+code.Host+`'`))
	} else {
		fmt.Println(BLANKS, green(`Uploaded code package with id='`+code.Id+`'`))
	}
	fmt.Println(BLANKS, green(u.hud_URL_str+"code/"+code.Id+INFO))
}

func (u *RegisterCmd) Flags(args ...string) error {
	u.flags = NewWorkerFlagSet()
	u.name = u.flags.name()
	u.maxConc = u.flags.maxConc()
	u.retries = u.flags.retries()
	u.retriesDelay = u.flags.retriesDelay()
	u.defaultPriority = u.flags.defaultPriority()
	u.config = u.flags.config()
	u.configFile = u.flags.configFile()
	u.envVars = u.flags.envVars()

	err := u.flags.Parse(args)
	if err != nil {
		return err
	}
	return u.flags.validateAllFlags()
}

// `iron worker register IMAGE`
func (u *RegisterCmd) Args() error {
	if u.flags.NArg() < 1 {
		return errors.New("command takes at least one argument. see -help")
	}

	u.codes.Command = strings.TrimSpace(strings.Join(u.flags.Args()[1:], " "))
	u.codes.Image = u.flags.Arg(0)

	u.codes.Name = u.codes.Image
	if strings.ContainsRune(u.codes.Name, ':') {
		arr := strings.SplitN(u.codes.Name, ":", 2)
		u.codes.Name = arr[0]
	}

	u.codes.MaxConcurrency = *u.maxConc
	u.codes.Retries = *u.retries
	u.codes.RetriesDelay = *u.retriesDelay
	u.codes.Config = *u.config
	u.codes.DefaultPriority = *u.defaultPriority

	if u.host != nil && *u.host != "" {
		u.codes.Host = *u.host
	}

	if *u.configFile != "" {
		pload, err := ioutil.ReadFile(*u.configFile)
		if err != nil {
			return err
		}
		u.codes.Config = string(pload)
	}

	if *u.envVars != nil {
		if envSlice, ok := u.envVars.Get().(envSlice); ok {
			envVarsMap := make(map[string]string, len(envSlice))
			for _, envItem := range envSlice {
				envVarsMap[envItem.Name] = envItem.Value
			}
			u.codes.EnvVars = envVarsMap
		}
	}

	return nil
}

func (u *RegisterCmd) Usage() {
	fmt.Fprintln(os.Stderr, `usage: iron worker register some/image[:tag]`)
	u.flags.PrintDefaults()
}

func (u *RegisterCmd) Run() {
	if u.codes.Host != "" {
		fmt.Println(LINES, `Spinning up '`+u.codes.Name+`'`)
	} else {
		fmt.Println(LINES, `Registering worker '`+u.codes.Name+`'`)
	}
	code, err := pushCodes("", &u.wrkr, u.codes)
	if err != nil {
		fmt.Println(err)
		return
	}
	if code.Host != "" {
		fmt.Println(BLANKS, green(`Hosted at: '`+code.Host+`'`))
	} else {
		fmt.Println(BLANKS, green(`Registered code package with id='`+code.Id+`'`))
	}
	fmt.Println(BLANKS, green(u.hud_URL_str+"code/"+code.Id+INFO))
}
