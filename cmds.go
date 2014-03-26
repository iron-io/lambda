package main

import (
	"errors"
	"fmt"
	"io/ioutil"
)

// TODO(reed): turn into client

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
	baseCmd
	codeName string
	payload  string
}

//type SchedCmd struct {
//// TODO(reed)
//baseCmd
//}

//type StatusCmd struct {
//// TODO(reed)
//baseCmd
//}

type LogCmd struct {
	// TODO(reed)
	baseCmd
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

// Takes one parameter, the task_id to log
func (q *QueueCmd) Args(args ...string) error {
	if len(args) != 1 {
		return errors.New("error: queue takes one argument")
	}
	q.codeName = args[0]

	// TODO(reed): this is bad, and I should feel bad
	if *payloadFileFlag != "" {
		body, err := ioutil.ReadFile(*payloadFileFlag)
		if err != nil {
			return err
		}
		q.payload = string(body)
		return nil
	}

	q.payload = *payloadFlag
	return nil
}

func (q *QueueCmd) Help() string {
	return `iron_worker [flags] queue $WORKER

[flags]:
-payload      = payload, typically json, to send to worker
-payloadFile  = location of file with payload
`
}

func (q *QueueCmd) Run() {
	// TODO(reed): delay, timeout, priority... move to Args()?
	task := struct {
		CodeName string `json:"code_name"`
		Payload  string `json:"payload"`
	}{
		q.codeName,
		q.payload,
	}

	resp, err := q.postJSON("/projects/"+q.ProjectID+"/tasks",
		map[string]interface{}{
			"tasks": []interface{}{task},
		})
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("%s", resp)
}

//func (s *SchedCmd) Help() string {
//return `
//iron_worker [flags] upload [.worker]

//[flags]:
//-imaflag
//`
//}

//func (s *SchedCmd) Run() {
//}

//func (s *StatusCmd) Help() string {
//return `
//iron_worker [flags] upload [.worker]

//[flags]:
//-imaflag
//`
//}

//func (s *StatusCmd) Run() {
//}

// Takes one parameter, the task_id to log
func (l *LogCmd) Args(args ...string) error {
	if len(args) != 1 {
		return errors.New("error: log takes one argument")
	}
	l.taskID = args[0]
	return nil
}

func (s *LogCmd) Help() string {
	return `iron_worker [flags] log task_id

[flags]:
-imaflag
`
}

func (s *LogCmd) Run() {
	log, err := s.get("/projects/" + s.ProjectID + "/tasks/" + s.taskID + "/log")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%s", log)
}
