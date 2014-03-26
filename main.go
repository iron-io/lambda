package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	payloadFlag     = flag.String("payload", "", "give worker payload")
	payloadFileFlag = flag.String("payload-file", "", "give worker payload file")
	helpFlag        = flag.Bool("h", false, "show this")
	commands        map[string]Command
)

func usage() {
	fmt.Fprintln(os.Stderr, `
  usage of iron_worker:

  iron_worker [flags] [command]
  
  run iron_worker -h [command] for [command] specific help
  `)
	os.Exit(0)
}

func init() {
	// TODO(reed): move this into getCommand()
	commands = map[string]Command{
		//"upload":   new(UploadCmd),
		//"run":      new(RunCmd),
		"queue": new(QueueCmd),
		//"schedule": new(SchedCmd),
		//"status":   new(StatusCmd),
		"log": new(LogCmd),
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

	//settings := config.Config("iron_worker")

	if err := cmd.Args(flag.Args()[1:]...); err != nil {
		fmt.Println(err)
		cmd.Help()
	}

	cmd.Config() // TODO(reed): this could be errors?
	cmd.Run()    // TODO(reed): want output?
}
