# IronCLI

Go version of the Iron.io command line tools.

## Install

### Quick and Easy (Recommended)

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

### See [other installation methods](#other-installation-methods) for more options.

## Getting Started

#### Before Getting Started

Before you can use IronWorker, be sure you've [created a free account with
Iron.io](http://www.iron.io) and [setup your Iron.io credentials on your
system](http://dev.iron.io/worker/reference/configuration/) (either in a json
file or using ENV variables). You only need to do that once for your machine. If
you've done that, then you can continue.

[See the official docs](http://dev.iron.io/worker/cli/) for more detailed info on using Docker for IronWorker.

#### Actually Getting Started

The easiest way to get started is by digging around.

`$ iron --help` for example usage and a list of commands

## Contributing

Give us a pull request! File a bug!

Since go1.5, we are lab rats in the go1.5 vendoring experiment. This eliminates
the need to modify import paths and depend on package maintainers not to break things.
For more info, see: <https://golang.org/s/go15vendor>.

Just `export GO15VENDOREXPERIMENT=1` in your shell env however you like, and
then `go build`.

## Other Installation Methods

### Download Yourself

Grab the latest version for your system on the [Releases](https://github.com/iron-io/ironcli/releases) page.

You can either run the binary directly or add somewhere in your $PATH.

### Use the iron/cli Docker image

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

