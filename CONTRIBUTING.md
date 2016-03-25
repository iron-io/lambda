Contributing to Lambda
======================

### Prerequisites

This has only been tested on OSX and Linux.

* Working Go 1.5 onwards installation.
* Working [Docker](http://docker.com) installation.
* GNU Make.

To work on Java code, you'll need a working JDK and [Apache Maven](http://maven.apache.org).
For node.js, any version of node >=0.10.0 will do.
For Python, Python 2.7 onwards.

### Workflow

Filing an issue on GitHub is a good first step. It establishes that something
is lacking and let's others know that you are working on it.

If it is a bug, please provide a Steps To Reproduce. There are several moving
parts in this project and having all the details helps:

* Your operating system (`uname -a` on UNIX systems).
* Versions of Go (`go version`) and Docker (`docker -v`).
* If you are using the Ironcli binary, `ironcli -version`.
* Base image version from Docker Hub (usually `latest`).
* The example Lambda Function and payload that caused the bug.

If you built a base image from source:
* Git SHA1 of the commit of this repository you used when the bug occurred.

If you are interested in fixing the bug or adding the missing feature, assign
the issue to yourself. If you plan to add some major feature, we recommend
mentioning one of the maintainers first to discuss.

The rest of the workflow is conventional Git practice.

* Fork the repository.
* Create a new branch - `git checkout -b my-feature`.
* Hack hack hack.
* Commit.
* `git push origin my-feature`
* File a Pull Request against upstream.

It is recommended to have one of your commits have `fixes #N` in the commit
message so that the issue you filed above is automatically closed when the PR
is accepted.

### Hacking on the images

Lambda works by having base Docker images for each platform. These provide
bootstrap scripts that create a Lambda like environment and provide the Lambda
APIs. Lambda functions created by users are simply auto-generated Docker images
layered onto these base images.

Images are located in `./images/<platform>` and have a Makefile to package the
bootstrap and create the image. Note that the Node.js and Python scripts are
not compiled or linted in any form, so make sure you test it out.

All changes that affect the Lambda environment should include tests added to
the test-suite. The test-suite README has comprehensive instructions on how to
add and run tests.

### Hacking on the Lambda workflow.

Creating Lambda functions is done by using the ironcli tool Lambda subcommands
(`ironcli lambda -help`). ironcli uses the `./lambda` library in this
repository under the hood. To improve this workflow, make sure your copy of
ironcli is built from source. To make sure you can hack on the `lambda` package
within the vendoring workflow, you should do something like this.

* Clone ironcli.
* Fork this repository.
* Edit the glide.yaml file to change the `lambda` package import path from
  `github.com/iron-io/lambda/lambda` to `github.com/<username>/lambda/lambda`.
  You'll also need to fix the imports in the ironcli lambda.go file.
* Run `glide i` to get the new package.
* Now `ironcli/vendor/github.com/<username>/lambda` will have your fork. Hack
  in here and submit a PR.
* If you have a better idea of a workflow, we would appreciate leaving us
  a note.
