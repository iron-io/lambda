package main

import (
	"errors"
	"fmt"
	"github.com/iron-io/iron_go/config"
	"github.com/iron-io/iron_go/mq"
	"os"
)

// IronMq v1

// type Command interface {
// 	Flags(...string) error // parse subcommand specific flags
// 	Args() error           // validate arguments
// 	Config() error         // configure env variables
// 	Usage() func()         // custom command help TODO(reed): all local now?
// 	Run()                  // cmd specific
// }

// It'd be better to abstract this out into two files, worker_command and mq_commands

type mqCommand struct {
	settings  config.Settings
	flags     *MqFlags
	token     *string
	projectID *string
}

func (mc *mqCommand) Config() error {
	mc.settings = config.ConfigWithEnv("iron_mq", *envFlag)

	if *projectIDFlag != "" {
		mc.settings.ProjectId = *projectIDFlag
	}
	if *tokenFlag != "" {
		mc.settings.Token = *tokenFlag
	}

	if mc.settings.ProjectId == "" {
		return errors.New("did not find project id in any config files or env variables")
	}
	if mc.settings.Token == "" {
		return errors.New("did not find token in any config files or env variables")
	}

	return nil
}

type ListCommand struct {
	mqCommand

	//flags
	page    *int
	perPage *int
}

func (l *ListCommand) Flags(args ...string) error {
	l.flags = NewMqFlagSet(l.Usage())

	l.page = l.flags.page()
	l.perPage = l.flags.perPage()

	err := l.flags.Parse(args)
	if err != nil {
		return err
	}
	return l.flags.validateAllFlags()
}

func (l *ListCommand) Args() error {
	return nil
}

func (l *ListCommand) Usage() func() {
	return func() {
		fmt.Fprintln(os.Stderr, `usage: iron mq list [--perPage perPpage] [--page page]
    --perPage perPage: Amount of queues showed per page
    --page page: starting page number`)
		return
	}
}

func (l *ListCommand) Run() {
	queues, err := mq.ListSettingsQueues(l.settings, 0, 30)
	if err != nil {
		fmt.Println(BLANKS, err)
		return
	}
	for _, q := range queues {
		fmt.Println(q.Name)
	}
}
