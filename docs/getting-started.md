# Introduction

This guide will walk you through creating and testing a simple Lambda function.

We will then upload it to IronWorker and run it.

We need the the `ironcli` tool for the rest of this guide. You can install it
by following [these instructions](https://github.com/iron-io/ironcli).

## Creating the function

Let's convert the `node-exec` AWS Lambda example to Docker. This simply
executes the command passed to it as the payload and logs the output.

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

Create an empty directory for your project and save this code in a file called
`node_exec.js`.

Now let's use `ironcli`'s Lambda functionality to create a Docker image. We can
then run the Docker image with a payload to execute the Lambda function.

    $ iron lambda create-function --function-name irontest/node-exec:1 --runtime nodejs --handler node_exec.handler node_exec.js
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
`1` indicates the version. You can use any string. This way you can configure
your deployment environment to use different versions. The handler is
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

The `test-function` subcommand can launch the Dockerized function with the
right parameters.

    $ iron lambda test-function --function-name irontest/node-exec:1 --payload '{ "cmd": "echo Dockerized Lambda" }'
    Dockerized Lambda!

You should see the output. Try changing the command to `date` or something more
useful.

## Uploading the function

You can run the Docker image anywhere. You can plug it into your orchestration
framework to launch it based on events. The ironcli tool allows publishing the
function directly to the IronWorker platform, where you can run it at scale,
without having to deal with machines and uptime.

[Sign up](signup) for a free IronWorker account. Follow the [IronWorker
guide][iwguide] to get your system ready. This means:

- You should have a Docker Hub ID and set up the credentials using `docker
  login`. In this tutorial, we'll use `irontest` as the Docker ID.
- Your environment should be [set up][iron-vars] with the credentials for your Iron.io account.

[signup]: http://www.iron.io/get-started/#start-trial
[iwguide]: http://dev.iron.io/worker/getting_started/
[iron-vars]: http://dev.iron.io/worker/reference/configuration/

The `publish-function` command first uploads the image to Docker Hub, then
registers it with IronWorker so you can queue tasks or launch a task in
response to a webhook.

    $ iron lambda publish-function --function-name irontest/node-exec:1
    ----->  Configuring client
            Project '<project name>' with id='<project id>'
    ----->  Registering worker 'irontest/node-exec'
            Registered code package with id='<id>'
            Check https://hud.iron.io/tq/projects/<project id>/code/<id> for more info

If you check [HUD](https://hud.iron.io), you should see your image on the Codes
page.

## Running the Lambda function

We can now run this from the command line.

    $ iron worker queue -payload '{ "cmd": "echo Dockerized Lambda" }' -wait irontest/node-exec
    ----->  Configuring client
            Project '<project name>' with id='<project id>'
    ----->  Queueing task 'irontest/node-exec'
            Queued task with id='<task id>'
            Check https://hud.iron.io/tq/projects/<project id>/jobs/<task id> for more info
    ----->  Waiting for task <task id>
    ----->  Done
    ----->  Printing Log:
    Dockerized Lambda!

IronWorker tasks are launched asynchronously. The `-wait` flag forces ironcli
to wait until the task is finished. The first run takes some time as
IronWorker has to fetch the Docker image. Subsequent runs are faster.

You can also launch the task via Webhooks. You can find the Webhook URL on the
Code page.

    $ curl -X POST -d '{ "cmd": "echo Dockerized Lambda" }' '<webhook URL>'

