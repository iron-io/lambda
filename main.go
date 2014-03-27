package main

import (
	"flag"
	"fmt"
	"os"
)

// TODO(reed): can flags be put in their respective type and parsed at runtime?
// this is ghastly and each method in cmds.go has flag pointers everywhere
var (
	// for queue
	payloadFlag     = flag.String("payload", "", "give worker payload")
	payloadFileFlag = flag.String("payload-file", "", "give worker payload file")
	priorityFlag    = flag.Int("priority", 0, "0(default), 1, or 2")
	timeoutFlag     = flag.Int("timeout", 3600, "0-3600(default) max runtime for task")
	delayFlag       = flag.Int("delay", 0, "seconds to delay before queueing task")
	waitFlag        = flag.Bool("wait", false, "wait for task to complete and print log")

	// for ---
	// ...

	helpFlag = flag.Bool("h", false, "show this")
	commands map[string]Command
)

const (
	LINES  = "-----> "
	BLANKS = "       "
	INFO   = "' for more info"
)

func usage() {
	fmt.Fprintln(os.Stderr, `usage of ironcli:

ironcli [flags] [command]

run 'ironcli -h [command]' for [command] specific help

[command]:`)
	for c, _ := range commands {
		fmt.Fprintln(os.Stderr, "\t", c)
	}
	fmt.Fprintln(os.Stderr, `[flag]:`)
	flag.PrintDefaults()
	os.Exit(0)
}

func init() {
	// TODO(reed): move this into getCommand()
	commands = map[string]Command{
		//"upload":   new(UploadCmd),
		//"run":      new(RunCmd),
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

	if *helpFlag {
		fmt.Println("usage for ", flag.Arg(0)+":")
		fmt.Printf("%s", cmd.Help())
		os.Exit(0)
	}

	if err := cmd.Args(flag.Args()[1:]...); err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	cmd.Config() // TODO(reed): this could be errors?
	cmd.Run()    // TODO(reed): want output?
}
