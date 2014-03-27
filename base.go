package main

import (
	"fmt"

	"github.com/iron-io/iron_go/config"
	"github.com/iron-io/iron_go/worker"
)

// TODO(reed): default flags for everybody
//--config CONFIG              config file
//-e, --env ENV                    environment
//--project-id PROJECT_ID      project id
//--token TOKEN                token

// The idea is:
//  validate arguments
//  if ^ goes well, config
//  if ^ goes well, run
//
//  ...and if anything goes wrong, help()
type Command interface {
	Args(...string) error // validate arguments
	Config()              // configure env variables
	Help() string         // custom command help, TODO(reed): export? really?
	Run()                 // cmd specific
}

// A command is mostly an alias for a worker, defining
// a Config() method that allows it to be an iron_worker specific worker
type command struct {
	worker.Worker
	hud_URL_str string
}

func (bc *command) Config() {
	bc.Settings = config.Config("iron_worker")
	bc.hud_URL_str = `Check 'http://hud.iron.io/tq/projects/` + bc.Settings.ProjectId + "/"
	fmt.Println(LINES, `Configuring client`)
	fmt.Println(BLANKS, `Project id="`+bc.Settings.ProjectId+`"`)
}
