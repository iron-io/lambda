package main

import (
	"errors"
	"fmt"

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
