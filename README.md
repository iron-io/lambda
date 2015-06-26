ironcli
=======

Go version of the Iron.io command line tools.  

# Install

## Quick and Easy (Recommended)

`curl -sSL http://get.iron.io/cli | sh`

## Download Yourself

Grab the latest version for your system on the [Releases](https://github.com/iron-io/ironcli/releases) page. 

You can either run the binary directly or add somewhere in your $PATH. 

## Coming soon...

Homebrew/deb/msi installer coming...

# Before Getting Started

Before you can use IronWorker, be sure you've [created a free account with
Iron.io](http://www.iron.io)
and [setup your Iron.io credentials on your
system](http://dev.iron.io/worker/reference/configuration/) (either in a json
file or using ENV variables). You only need to do that once for your machine. If
you've done that, then you can continue.

[See the official docs](http://dev.iron.io/worker/beta/cli/) for more detailed info on using Docker for IronWorker.

# Help

`iron worker --help` for list of commands, flags
`iron worker COMMAND --help` for command specific help

# Currently supported commands

__WARNING:__ still in progress (only upload problematic), if running into issues: use `github.com/iron-io/iron_worker_ruby_ng`

### Queue a task: 

`iron worker queue CODENAME`

### Wait for queued task and print log: 

`iron worker queue --wait CODENAME`

### Status of task:

`iron worker status TASK_ID`

Hint: Acquire `TASK_ID` from a previously queued task.

### Log task:

`iron worker log TASK_ID`

Hint: Acquire `TASK_ID` from a previously queued task.

### Schedule task:

`iron worker schedule --payload=" " --start-at="Mon Dec 25 15:04:05 -0700 2014" CODENAME`

__WARNING:__ not working without a -payload for reasons yet to be hunted down

### Upload a worker:

`iron worker upload [--zip hello.zip] --name NAME DOCKER_IMAGE [COMMAND]`

For eg:

`iron worker upload --zip myworker.zip --name myworker iron/images:ruby-2.1 ruby hello.rb`

For custom images (if you have this enabled on your account): 

`iron worker upload --zip myworker.zip --name myworker google/ruby ruby hello.rb`

## Contributing

Give us a pull request!

Updated dependencies:

* `go get -u yourlib`
* `godep update yourlib`
* `godep save -r`

