// Package contains the command line interface for iron-worker.
package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	// These are located after binary on command line
	// TODO(reed): kind of awkward, since there are 2 different flag sets now:
	//  e.g.
	//    ironcli -token=123456789 upload -max-concurrency=10 my_worker
	helpFlag      = flag.Bool("help", false, "show this")
	tokenFlag     = flag.String("token", "", "provide OAuth token")
	projectIDFlag = flag.String("project-id", "", "provide project ID")
	envFlag       = flag.String("env", "", "provide specific dev environment")

	commands map[string]Command
)

const (
	LINES  = "-----> "
	BLANKS = "       "
	INFO   = "' for more info"
)

func usage() {
	fmt.Fprintln(os.Stderr, `usage of ironcli: ironcli [command] [flags] [args]

  run 'ironcli -help [command]' for [command]'s flags/args

[command]:`)
	for c, _ := range commands {
		fmt.Fprintln(os.Stderr, "\t", c)
	}
	fmt.Fprintln(os.Stderr, `[flag]:`)
	flag.PrintDefaults()
	os.Exit(0)
}

func init() {
	commands = map[string]Command{
		"upload":   new(UploadCmd),
		"run":      new(RunCmd),
		"queue":    new(QueueCmd),
		"schedule": new(SchedCmd),
		"status":   new(StatusCmd),
		"log":      new(LogCmd),
	}
}

func main() {
	flag.Parse()

	if flag.NArg() < 1 {
		usage()
	}

	cmd, ok := commands[flag.Arg(0)]

	if !ok {
		fmt.Fprintln(os.Stderr, flag.Arg(0), "not a command, see -h")
		os.Exit(0)
	}

	// each command defines its flags, err is either ErrHelp or bad flag value
	if err := cmd.Flags(flag.Args()[1:]...); err != nil {
		if err != flag.ErrHelp {
			fmt.Println(err)
		}
		os.Exit(2)
	}

	if err := cmd.Args(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	err := cmd.Config()
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	cmd.Run()
}
