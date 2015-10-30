ironcli
=======

Go version of the Iron.io command line tools.

# Install

## Quick and Easy (Recommended)

`curl -sSL https://cli.iron.io/install | sh`

If you're concerned about the [potential insecurity](http://curlpipesh.tumblr.com/)
of using `curl | sh`, feel free to use a two-step version of our installation and examine our
installation script:

```bash
curl -f -sSL https://cli.iron.io/install -O
sh install
```

#### `curl | sh` as an Installation Method?

We'd like to explain why we're telling you to `curl | sh` to install this software.
The script at https://cli.iron.io/install has some relatively simple logic to download the
right `ironcli` binary for your platform. When you run that script by piping the `curl` output
into `sh`, you're trusting us that it's safe and won't harm your computer. We hope that you do!
But if you don't please see the section just below this one on how to download and run the binary
yourself without an install script.

## Download Yourself

Grab the latest version for your system on the [Releases](https://github.com/iron-io/ironcli/releases) page.

You can either run the binary directly or add somewhere in your $PATH.

## Use the iron/cli Docker image

If you have Docker installed, then you don't need to install anything else to use this.
All the commands are the same, but instead of starting the command with `iron`, change it to:

```sh
docker run --rm -it -v "$PWD":/app -w /app iron/cli ...
```

If you're using the Docker image, you either need to have your `iron.json` file in the local directory (it won't pick it up from $HOME),
or set your Iron credentials in environment variables:

```sh
export IRON_TOKEN=YOURTOKEN
export IRON_PROJECT_ID=YOURPROJECT_ID
```

And then use `-e` flags with the docker run command:

```sh
docker run --rm -it -e IRON_TOKEN -e IRON_PROJECT_ID -v "$PWD":/app -w /app iron/cli ...
```

# Before Getting Started

Before you can use IronWorker, be sure you've [created a free account with
Iron.io](http://www.iron.io)
and [setup your Iron.io credentials on your
system](http://dev.iron.io/worker/reference/configuration/) (either in a json
file or using ENV variables). You only need to do that once for your machine. If
you've done that, then you can continue.

[See the official docs](http://dev.iron.io/worker/cli/) for more detailed info on using Docker for IronWorker.

# Help

- `iron worker --help` for list of commands and flags
- `iron worker COMMAND --help` for command-specific help

# Currently supported commands

__WARNING:__ still in progress (only upload problematic), if running into
issues, try falling back to [the previous generation Ruby
CLI](https://github.com/iron-io/iron_worker_ruby_ng) (sorry!).

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
## IronMQ

### List the queues in a project

`iron mq list`

### Create a new queue
`iron mq create QUEUE_NAME`

### Delete an existing queue
`iron mq rm QUEUE_NAME`

### Display a queue's details

`iron mq info QUEUE_NAME`

### Clear all message on a queue

`iron mq clear QUEUE_NAME`

### Push a message to a queue

`iron mq push [-f file] QUEUE_NAME "MESSAGE"`

You can provide a json file with a set of messages to be pushed onto the queue. The format is as follows:
```json
{
  "messages": ["msg1", "msg2",...]
}
```
### Peek n message from a queue
`iron mq peek [-n n] QUEUE_NAME`

### Pop (get and delete) a set of messages from the queue
`iron mq pop [-o output_file] [-n n] QUEUE_NAME`

### Reserve a set of messages from a queue
`iron mq reserve [-o output_file] [-n n] [-t timeout] QUEUE_NAME`

### Delete a set of reserved message
`iron mq delete [-f file] QUEUE_NAME "MESSAGE_ID" "MESSAGE_ID2"...`

For private images you should use
`iron worker docker-login --repo-username USERNAME --repo-pass PASS  --repo-email EMAIL`
Or
`iron worker docker-login --repo-auth AUTH --repo-email EMAIL`

## Contributing

Give us a pull request!

Updated dependencies:

Until the go team standardizes on a vendoring tool, we're using something
analogous to their proposed plan for 1.5, to update a dependency, see:

[vendoring](vendored/README.md)
