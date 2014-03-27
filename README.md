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

Since our worker will be executed in the cloud, you'll need to bundle all the
necessary gems,
supplementary data, and other dependencies with it. `.worker` files make it easy
to define your worker.

```ruby
# define the runtime language, this can be ruby, java, node, php, go, etc.
runtime "ruby"
# exec is the file that will be executed:
exec "hello_worker.rb"
```

You can read more about `.worker` files here:
http://dev.iron.io/worker/reference/dotworker/


# Help

`ironcli -h` for list of commands, flags
`ironcli -h COMMAND` for command specific help

__WARNING:__ flags are deceiving and currently only work for `queue` except `-h`

These are going to change in the very near future to be subcommand based flags
that appear after the COMMAND... so don't write any scripts just yet.

# Currently supported commands

__WARNING:__ cannot upload from `ironcli` currently, use `github.com/iron-io/iron_worker_ruby_ng`

### Queue a task: 

`ironcli queue CODENAME`

Currently working flags:

* `-wait`
* `-priority`
* `-payload`
* `-payload-file`
* `-delay`
* `-timeout`

### Wait for queued task and print log: 

`ironcli -wait queue CODENAME`

### Status of task:

`ironcli status TASK_ID`

Hint: Acquire `TASK_ID` from a previously queued task.

### Log task:

`ironcli log TASK_ID`

Hint: Acquire `TASK_ID` from a previously queued task.

### Schedule task:

`ironcli schedule CODENAME`

__WARNING:__ not working without a -payload for reasons yet to be hunted down

... currently not very useful because the actual scheduling flags aren't
implemented, it's basically an alias for queue as it stands

