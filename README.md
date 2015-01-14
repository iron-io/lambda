ironcli
=======

Go version of the Iron.io command line tools.  

# Install

## Quick and Easy (Recommended)

`curl -sSL http://get.iron.io/cli | sh`

## Download Yourself

__Downloads (all 64-bit):__

[Mac](https://github.com/iron-io/ironcli/releases/download/v0.0.2-alpha/ironcli_mac)

[Linux](https://github.com/iron-io/ironcli/releases/download/v0.0.2-alpha/ironcli_linux)

[Windows](https://github.com/iron-io/ironcli/releases/download/v0.0.2-alpha/ironcli.exe)

You can either run the binary directly or add somewhere in your $PATH

## Coming soon...

Homebrew/deb/msi installer coming...

# Before Getting Started

Before you can use IronWorker, be sure you've [created a free account with
Iron.io](http://www.iron.io)
and [setup your Iron.io credentials on your
system](http://dev.iron.io/worker/reference/configuration/) (either in a json
file or using ENV variables). You only need to do that once for your machine. If
you've done that, then you can continue.

# Help

`ironcli --help` for list of commands, flags
`ironcli COMMAND --help` for command specific help

# Currently supported commands

__WARNING:__ still in progress (only upload problematic), if running into issues: use `github.com/iron-io/iron_worker_ruby_ng`

### Queue a task: 

`ironcli queue CODENAME`

### Wait for queued task and print log: 

`ironcli queue --wait CODENAME`

### Status of task:

`ironcli status TASK_ID`

Hint: Acquire `TASK_ID` from a previously queued task.

### Log task:

`ironcli log TASK_ID`

Hint: Acquire `TASK_ID` from a previously queued task.

### Schedule task:

`ironcli schedule --payload=" " --start-at="Mon Dec 25 15:04:05 -0700 2014" CODENAME`

__WARNING:__ not working without a -payload for reasons yet to be hunted down

### Upload a task:

With a .worker in current directory (see /test\_worker):

`ironcli upload WORKER`

Currently, runtime specific dependencies are not supported. "file" and "dir"
should work fine. "full\_remote\_build" also not yet supported.

### Run a task locally (useful for testing):

`ironcli run WORKER`

Where WORKER is the name of a .worker file.

Due to environment disparities, this could not work on a local environment yet
work fine once uploaded. Things should work best under amd64 Ubuntu Linux.
