# Introduction

This guide will walk you through running a simple Lambda function on
IronWorker.

## Prerequisites

Sign up for a free IronWorker account. Follow the [IronWorker guide][iwguide]
to get your system ready. This means:

- You should have a working [Go][go] >=1.5 installation.
- You should have [Docker][docker] installed and working.
- You should have a Docker Hub ID and set up the credentials using `docker
  login`. In this tutorial, we'll use `irontest` as the Docker ID.
- You should have [environment variables][iron-vars] set to interact with IronWorker through
  ironcli.

[iwguide]: http://dev.iron.io/worker/getting_started/
[go]: http://golang.org
[docker]: http://www.docker.com
[iron-vars]: http://dev.iron.io/worker/reference/configuration/

Until the Lambda Docker images are published on Docker Hub, you'll also need to
build those yourselves. If you haven't already cloned this repository, do so
now.

    $ git clone https://github.com/iron-io/lambda
    $ cd lambda/images/node
    $ make

This will create a base Docker image `iron/lambda-nodejs`.

We are going to use a development branch of `ironcli` instead of the official
release.

    $ cd $GOPATH
    $ mkdir -p src/github.com/iron-io
    $ cd src/github.com/iron-io
    $ git clone https://github.com/iron-io/ironcli
    $ cd ironcli
    $ git checkout -b lambda -t origin/lambda
    $ go install .

## Creating the function

Let's convert the `node-exec` lambda example to Docker.

    var exec = require('child_process').exec;
    
    exports.handler = function(event, context) {
        if (!event.cmd) {
            context.fail('Please specify a command to run as event.cmd');
            return;
        }
        var child = exec(event.cmd, function(error) {
            // Resolve with result of process
            context.done(error, 'Process complete!');
        });
    
        // Log process stdout and stderr
        child.stdout.on('data', console.log);
        child.stderr.on('data', console.error);
    };

Create an empty directory for your project and save this code in a file called `node_exec.js`.

Now let's use `ironcli`'s Lambda functionality to create a Docker image. We can
then run the Docker image with a payload to execute the Lambda function.

    $ $GOPATH/bin/ironcli lambda create-function --function-name irontest/node-exec:1 --runtime nodejs --handler node_exec.handler node_exec.js
    Image output Step 1 : FROM iron/lambda-nodejs
    ---> 66fb7af42230
    Step 2 : ADD node_exec.js ./node_exec.js
    ---> 6f922128da71
    Removing intermediate container 9644b02e95bc
    Step 3 : CMD node_exec.handler
    ---> Running in 47b2b1f3e779
    ---> 5eef8d2d3111
    Removing intermediate container 47b2b1f3e779
    Successfully built 5eef8d2d3111

As you can see, this is very similar to creating a Lambda function using the
`aws` CLI tool. We name the function as we would name other Docker images. The
`1` indicates the version. You can use any string. This way you can make
changes to your code and tell IronWorker to run the newer code. The handler is
the name of the function to run, in the form that nodejs expects
(`module.function`). Where you would package the files into a `.zip` to upload
to Lambda, we just pass the list of files to `ironcli`. If you had node
dependencies you could pass the `node_modules` folder too.

You should now see the generated Docker image.

    $ docker images
    REPOSITORY                                      TAG    IMAGE ID         CREATED             VIRTUAL SIZE
    irontest/node-exec                              1      5eef8d2d3111     9 seconds ago       44.94 MB
    ...

## Testing the function

Testing your function locally is a little involved at this point. Eventually it
will be done through `ironcli`.

IronWorker expects the payload to be in a file defined by the environment
variable `PAYLOAD_FILE`. Let's create the JSON payload. Our function expects
a key called `cmd`. Call the file `payload.json`:

    {
      "cmd": "echo 'Dockerized Lambda!'"
    }

We will run our Docker image with volume sharing to allow it to read the
payload.

    $ docker run --rm -v $PWD:/mnt -e PAYLOAD_FILE=/mnt/payload.json irontest/node-exec:1
    Dockerized Lambda!

You should see the output. Try changing the command.

## Uploading the function

Uploading is a two stage process. First we upload the image to Docker Hub so
IronWorker can use it.

    $ docker push irontest/node-exec:1

Then we tell IronWorker to use this image as our lambda function.

    $ $GOPATH/bin/ironcli register 
    ----->  Configuring client
            Project '<project name>' with id='<project id>'
    ----->  Registering worker 'irontest/node-exec'
            Registered code package with id='<id>'
            Check https://hud.iron.io/tq/projects/<project id>/code/<id> for more info

If you check [HUD](https://hud.iron.io), you should see your image on the Codes
page.

## Running the Lambda function

We can now run this from the command line.

    $ $GOPATH/bin/ironcli worker queue -payload-file payload.json -wait irontest/node-exec
    ----->  Configuring client
            Project '<project name>' with id='<project id>'
    ----->  Queueing task 'irontest/node-exec'
            Queued task with id='<task id>'
            Check https://hud.iron.io/tq/projects/<project id>/jobs/<task id> for more info
    ----->  Waiting for task <task id>
    ----->  Done
    ----->  Printing Log:
    Dockerized Lambda!

The first run takes some time as IronWorker has to fetch the Docker image.
Subsequent runs are faster.

You can also launch the task via Webhooks. You can find the Webhook URL on the
Code page.

    $ curl -X POST -d @payload.json '<webhook URL>'

---

TODO:
Add test-function and publish-function commands to CLI and then replace the
Upload and Test manual steps with them.
NOTE to devs: Perhaps prefetch Lambda function runners on to certain machines?
