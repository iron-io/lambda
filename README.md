ironcli
=======

Go version of the cli. 

WIP

# Install

`go get github.com/iron-io/ironcli`

# Before Getting Started

Before you can use IronWorker, be sure you've [created a free account with
Iron.io](http://www.iron.io)
and [setup your Iron.io credentials on your
system](http://dev.iron.io/worker/reference/configuration/) (either in a json
file or using ENV variables). You only need to do that once for your machine. If
you've done that, then you can continue.

# Help

`ironcli -help` for list of commands, flags
`ironcli COMMAND -help` for command specific help

# Currently supported commands

__WARNING:__ cannot upload most things from `ironcli` currently, use `github.com/iron-io/iron_worker_ruby_ng`

But a basic one file "go" or "binary" worker currently works, if you want to do that.

### Queue a task: 

`ironcli queue CODENAME`

### Wait for queued task and print log: 

`ironcli queue -wait CODENAME`

### Status of task:

`ironcli status TASK_ID`

Hint: Acquire `TASK_ID` from a previously queued task.

### Log task:

`ironcli log TASK_ID`

Hint: Acquire `TASK_ID` from a previously queued task.

### Schedule task:

`ironcli schedule -payload=" " -start-at="Mon Dec 25 15:04:05 -0700 2014" CODENAME`

__WARNING:__ not working without a -payload for reasons yet to be hunted down

### Upload a task:

With a (currently basic) .worker in current directory (see /test\_worker):

`ironcli upload WORKER`
