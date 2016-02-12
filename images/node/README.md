Support for running nodejs Lambda functions.

Create an image with

    docker build -t iron/node-lambda .

This sets up a node stack and installs some deps to provide the Lambda runtime.

Right now we use [node-lambda](https://github.com/motdotla/node-lambda) to exec
the handler and provide the function. Unfortunately it requires
a `package.json`, hence the `package.json.stupid`. Eventually we should break
the `node-lambda run` part of it out and use that.

Running
-------

Does not support payload (AWS Lambda 'event') yet. Expects the lambda zip file
to be called `function.zip` and available at `/mnt/function.zip`. HANDLER
should be set to `module.*export*`. So 

    // fancy.js inside function.zip
    exports.fancyFunction = function(event, context) {}

would be launched as:

    docker run --rm -it -v /path/to/dir/containing/function.zip:/mnt -e HANDLER=fancy.fancyFunction iron/node-lambda

In Lambda you'd submit these parameters in the call to `create-function`.
