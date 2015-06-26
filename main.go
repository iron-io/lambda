// Package contains the command line interface for iron-worker.
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/iron-io/ironcli/Godeps/_workspace/src/github.com/fatih/color"
)

var (
	// These are located after binary on command line
	// TODO(reed): kind of awkward, since there are 2 different flag sets now:
	//  e.g.
	//    ironcli -token=123456789 upload -max-concurrency=10 my_worker
	versionFlag   = flag.Bool("version", false, "print then version number")
	helpFlag      = flag.Bool("help", false, "show this")
	hFlag         = flag.Bool("h", false, "show this")
	tokenFlag     = flag.String("token", "", "provide OAuth token")
	projectIDFlag = flag.String("project-id", "", "provide project ID")
	envFlag       = flag.String("env", "", "provide specific dev environment")

	red, yellow, green func(a ...interface{}) string

	// i.e. worker: { commands... }
	//			mq:			{ commands... }
	commands map[string]map[string]Command
)

const (
	LINES  = "-----> "
	BLANKS = "       "
	INFO   = " for more info"

	Version = "v0.0.14"
)

func usage() {
	fmt.Fprintln(os.Stderr, "usage: ", os.Args[0], `[product] [command] [flags] [args]

where [product] is one of:

  worker

run '`+os.Args[0], `[product] -help for a list of commands.
run '`+os.Args[0], `[product] [command] -help' for [command]'s flags/args.
`)
	fmt.Fprintln(os.Stderr, `[flags]:`)
	flag.PrintDefaults()
	os.Exit(0)
}

func pusage(p string) {
	// TODO list commands
	switch p {
	case "worker":
		fmt.Fprintln(os.Stderr, p, "commands:")
		for cmd := range commands["worker"] {
			fmt.Fprintln(os.Stderr, "\t", cmd)
		}
		os.Exit(0)
	case "mq":
		fmt.Fprintln(os.Stderr, "not yet")
		os.Exit(1)
	default:
		fmt.Fprintln(os.Stderr, red("invalid product ", `"`+p+`", `, "see -help"))
		os.Exit(1)
	}
}

func init() {
	commands = map[string]map[string]Command{
		"worker": map[string]Command{
			"upload":   new(UploadCmd),
			"queue":    new(QueueCmd),
			"schedule": new(SchedCmd),
			"status":   new(StatusCmd),
			"log":      new(LogCmd),
		},
		// TODO mq
	}
}

func main() {
	if runtime.GOOS == "windows" {
		red = fmt.Sprint
		yellow = fmt.Sprint
		green = fmt.Sprint
	} else {
		red = color.New(color.FgRed).SprintFunc()
		yellow = color.New(color.FgYellow).SprintFunc()
		green = color.New(color.FgGreen).SprintFunc()
	}

	flag.Parse()

	if *helpFlag || *hFlag {
		usage()
	} else if *versionFlag {
		fmt.Fprintln(os.Stderr, Version)
		os.Exit(0)
	}

	if flag.NArg() < 1 {
		usage()
	}

	product := flag.Arg(0)
	cmds, ok := commands[product]
	if !ok {
		pusage(product)
	}

	if flag.NArg() < 2 {
		pusage(product)
	}

	cmdName := flag.Arg(1)
	cmd, ok := cmds[cmdName]

	if !ok {
		switch strings.TrimSpace(cmdName) {
		case "-h", "help", "--help", "-help":
			pusage(product)
		default:
			fmt.Fprintln(os.Stderr, red(cmdName, " not a command, see -h"))
		}
		os.Exit(1)
	}

	// each command defines its flags, err is either ErrHelp or bad flag value
	if err := cmd.Flags(flag.Args()[2:]...); err != nil {
		if err != flag.ErrHelp {
			fmt.Println(red(err))
		}
		os.Exit(2)
	}

	if err := cmd.Args(); err != nil {
		fmt.Fprintln(os.Stderr, red(err))
		os.Exit(2)
	}

	err := cmd.Config()
	if err != nil {
		fmt.Println(red(err))
		os.Exit(2)
	}

	cmd.Run()
}
