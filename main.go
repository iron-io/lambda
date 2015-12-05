// Package contains the command line interface for iron-worker.
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/fatih/color"
)

var (
	// These are located after binary on command line
	// TODO(reed): kind of awkward, since there are 2 different flag sets now:
	//  e.g.
	//    ironcli -token=123456789 upload -max-concurrency=10 my_worker
	versionFlag   = flag.Bool("version", false, "print the version number")
	helpFlag      = flag.Bool("help", false, "show this")
	hFlag         = flag.Bool("h", false, "show this")
	tokenFlag     = flag.String("token", "", "provide OAuth token")
	projectIDFlag = flag.String("project-id", "", "provide project ID")
	envFlag       = flag.String("env", "", "provide specific dev environment")

	red, yellow, green func(a ...interface{}) string

	// i.e. worker: { commands... }
	//      mq:     { commands... }
	commands = map[string]commander{
		"run": runner{},
		"docker": mapper{
			"login": new(DockerLoginCmd),
		},
		"register": registrar{},
		"worker": mapper{
			"upload":   new(UploadCmd),
			"queue":    new(QueueCmd),
			"schedule": new(SchedCmd),
			"status":   new(StatusCmd),
			"log":      new(LogCmd),
		},
		"mq": mapper{
			"push":    new(PushCmd),
			"pop":     new(PopCmd),
			"reserve": new(ReserveCmd),
			"delete":  new(DeleteCmd),
			"peek":    new(PeekCmd),
			"clear":   new(ClearCmd),
			"list":    new(ListCmd),
			"create":  new(CreateCmd),
			"rm":      new(RmCmd),
			"info":    new(InfoCmd),
		},
	}
)

const (
	LINES  = "-----> "
	BLANKS = "       "
	INFO   = " for more info"

	Version = "0.0.21"
)

func usage() {
	fmt.Fprintln(os.Stderr, "usage: ", os.Args[0], `[product] [command] [flags] [args]

where [product] is one of:

  mq
  worker
  docker
  register
  run

run '`+os.Args[0], `[product] -help for a list of commands.
run '`+os.Args[0], `[product] [command] -help' for [command]'s flags/args.
`)
	fmt.Fprintln(os.Stderr, `[flags]:`)
	flag.PrintDefaults()
	os.Exit(0)
}

func pusage(p string) {
	prod, ok := commands[p]
	if !ok {
		fmt.Fprintln(os.Stderr, red("invalid product ", `"`+p+`", `, "see -help"))
		os.Exit(1)
	}
	fmt.Fprintln(os.Stderr, p, "commands:")
	for _, cmd := range prod.Commands() {
		fmt.Fprintln(os.Stderr, "\t", cmd)
	}
	os.Exit(0)
}

type commander interface {
	// Given a full set of command line args, call Args and Flags with
	// whatever position needed to be sufficiently rad.
	Command(args ...string) (Command, error)
	Commands() []string
}

type (
	// mapper expects > 0 args, calls flags after first arg
	mapper map[string]Command
	// runner calls flags on first (zeroeth) arg
	runner struct{}
	// registrar calls flags on first (zeroeth) arg, using RegisterCmd
	registrar struct{}
)

func (r runner) Commands() []string { return []string{"just run!"} } // --help handled in Flags()
func (r runner) Command(args ...string) (Command, error) {
	run := new(RunCmd)
	err := run.Flags(args[0:]...)
	if err == nil {
		err = run.Args()
	}
	return run, err
}

func (r registrar) Commands() []string { return []string{"just register!"} } // --help handled in Flags()
func (r registrar) Command(args ...string) (Command, error) {
	run := new(RegisterCmd)
	err := run.Flags(args[0:]...)
	if err == nil {
		err = run.Args()
	}
	return run, err
}

func (m mapper) Commands() []string {
	var c []string
	for cmd := range m {
		c = append(c, cmd)
	}
	return c
}

func (m mapper) Command(args ...string) (Command, error) {
	c, ok := m[args[0]]
	if !ok {
		return nil, fmt.Errorf("command not found: %s", args[0])
	}
	err := c.Flags(args[1:]...)
	if err == nil {
		err = c.Args()
	}
	return c, err
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
		fmt.Println(Version)
		os.Exit(0)
	}

	if flag.NArg() < 1 {
		usage()
	}

	product := flag.Arg(0)
	cmds, ok := commands[product]
	if !ok || flag.NArg() < 2 {
		pusage(product)
	}

	cmdName := flag.Arg(1)
	cmd, err := cmds.Command(flag.Args()[1:]...)

	if err != nil {
		if err == flag.ErrHelp && cmd != nil {
			cmd.Usage()
		}
		switch strings.TrimSpace(cmdName) {
		case "-h", "help", "--help", "-help":
			pusage(product)
		default:
			fmt.Fprintln(os.Stderr, red(err))
		}
		os.Exit(1)
	}

	err = cmd.Config()
	if err != nil {
		fmt.Fprintln(os.Stderr, red(err))
		os.Exit(2)
	}

	cmd.Run()
}
