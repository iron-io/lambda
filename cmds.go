package main

// TODO(reed): fix: empty schedule payload not working ?

import (
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/iron-io/iron_go/worker"
)

//type UploadCmd struct {
//baseCmd
//// TODO(reed)
//}

//type RunCmd struct {
//// TODO(reed)
//baseCmd
//}

type QueueCmd struct {
	// TODO(reed)
	command
	task worker.Task
}

type SchedCmd struct {
	// TODO(reed)
	command
	sched worker.Schedule
}

type StatusCmd struct {
	// TODO(reed)
	command
	taskID string
}

type LogCmd struct {
	// TODO(reed)
	command
	taskID string
}

//func (u *UploadCmd) Help() string {
//return `
//iron_worker [flags] upload [.worker]

//[flags]:
//-imaflag
//`
//}

//func (u *UploadCmd) Run() {
//}

//func (r *RunCmd) Help() string {
//return `
//iron_worker [flags] upload [.worker]

//[flags]:
//-imaflag
//`
//}

//func (r *RunCmd) Run() {
//}

// Takes one parameter, the code package to schedule
func (s *SchedCmd) Args(args ...string) error {
	if len(args) != 1 {
		return errors.New("error: queue takes one argument")
	}

	if err := checkPriority(); err != nil {
		return err
	}
	payload, err := payload()
	if err != nil {
		return err
	}

	delay := time.Duration(*delayFlag) * time.Second

	s.sched = worker.Schedule{
		CodeName: args[0],
		Payload:  payload,
		Delay:    &delay,
		Priority: priorityFlag,
	}

	return nil
}

// TODO(reed): move me
func payload() (string, error) {
	if *payloadFileFlag != "" {
		body, err := ioutil.ReadFile(*payloadFileFlag)
		return string(body), err
	}
	return *payloadFlag, nil
}

// TODO(reed): move me
func checkPriority() error {
	if *priorityFlag < 0 || *priorityFlag > 2 {
		return errors.New("priority can only be 0(default), 1, or 2")
	}
	return nil
}

func (q *SchedCmd) Help() string {
	return `usage: iron_worker schedule CODE_PACKAGE_NAME [OPTIONS]
-payload      = payload, typically json, to send to worker
-payloadFile  = location of file with payload
-p, --payload PAYLOAD            payload to pass
-f, --payload-file PAYLOAD_FILE  payload file to pass
--priority PRIORITY          0 (default), 1, 2
--timeout TIMEOUT            maximum run time in seconds from 0 to 3600 (default)
--delay DELAY                delay before start in seconds
--start-at TIME              start task at specified time
--end-at TIME                stop running task at specified time
--run-times RUN_TIMES        run task no more times than specified
--run-every RUN_EVERY        run task every RUN_EVERY seconds
`
}

func (s *SchedCmd) Run() {
	fmt.Println(LINES, "Scheduling task")

	ids, err := s.Schedule(s.sched)
	if err != nil {
		fmt.Println(BLANKS, err)
		return
	}
	// TODO(reed): > 1 ever?
	id := ids[0]

	fmt.Printf("%s scheduled %s with id: %s\n", BLANKS, s.sched.CodeName, id)
	fmt.Println(BLANKS, s.hud_URL_str+"scheduled_jobs/"+id, INFO)
}

// Takes one parameter, the codes name to queue
func (q *QueueCmd) Args(args ...string) error {
	// TODO(reed): delay, timeout, priority... move to Args()?
	if len(args) != 1 {
		return errors.New("error: queue takes one argument")
	}

	if err := checkPriority(); err != nil {
		return err
	}
	payload, err := payload()
	if err != nil {
		return err
	}

	delay := time.Duration(*delayFlag) * time.Second
	timeout := time.Duration(*timeoutFlag) * time.Second

	q.task = worker.Task{
		CodeName: args[0],
		Payload:  payload,
		Delay:    &delay,
		Timeout:  &timeout,
		Priority: *priorityFlag,
	}

	return nil
}

func (q *QueueCmd) Help() string {
	return `usage: iron_worker queue CODE_PACKAGE_NAME [OPTIONS]
-p, --payload PAYLOAD            payload to pass
-f, --payload-file PAYLOAD_FILE  payload file to pass
--priority PRIORITY          0 (default), 1, 2
--timeout TIMEOUT            maximum run time in seconds from 0 to 3600 (default)
--delay DELAY                delay before start in seconds
--wait                       wait for task to complete and print log
`
}

func (q *QueueCmd) Run() {
	fmt.Println(LINES, "Queueing task")

	ids, err := q.TaskQueue(q.task)
	if err != nil {
		fmt.Println(BLANKS, err)
		return
	}
	// TODO(reed): > 1 ever?
	id := ids[0]

	fmt.Printf("%s Queued %s with id: %s\n", BLANKS, q.task.CodeName, id)
	fmt.Println(BLANKS, q.hud_URL_str+"jobs/"+id+INFO)

	if *waitFlag {
		fmt.Println(LINES, "Waiting for task", id)

		out := q.WaitForTaskLog(id)

		log := <-out
		fmt.Println(LINES, "Done")
		fmt.Println(LINES, "Printing Log:")
		fmt.Printf("%s", string(log))
	}
}

// Takes one parameter, the task_id to acquire status of
func (s *StatusCmd) Args(args ...string) error {
	if len(args) != 1 {
		return errors.New("error: status takes one argument")
	}
	s.taskID = args[0]
	return nil
}

// TODO(reed): flags
func (s *StatusCmd) Help() string {
	return `iron_worker [flags] status task_id

[flags]:
-imaflag
`
}

func (s *StatusCmd) Run() {
	fmt.Println(LINES, "Getting status of task with id", s.taskID)
	taskInfo, err := s.TaskInfo(s.taskID)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(BLANKS, taskInfo.Status)
}

// Takes one parameter, the task_id to log
func (l *LogCmd) Args(args ...string) error {
	if len(args) != 1 {
		return errors.New("error: log takes one argument")
	}
	l.taskID = args[0]
	return nil
}

// TODO(reed): flags
func (l *LogCmd) Help() string {
	return `iron_worker [flags] log task_id

[flags]:
-imaflag
`
}

func (l *LogCmd) Run() {
	fmt.Println(LINES, "Getting log for task with id", l.taskID)
	out, err := l.TaskLog(l.taskID)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(out))
}
