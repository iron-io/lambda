# Lambda everywhere.

AWS Lambda introduced server-less computing to the masses. Wouldn't it be nice
if you could run the same Lambda functions on any platform, in any cloud?
Iron.io is proud to release a set of tools that allow just this. Package your
Lambda function in a Docker container and run it anywhere with the same
environment that AWS Lambda provides.
                          --payload '{ "cmd": "echo Dockerized Lambda" }'
    Dockerized Lambda

## Use cases

## How does it work?

We provide base Docker images for the various runtimes that AWS Lambda
supports. The `ironcli` tool helps package up your Lambda function into
a Docker image layered on the base image. We provide a bootstrap script and
utilities that provide a AWS Lambda environment to your code. You can then run
the Docker image on any platform that supports Docker. This allows you to
easily move Lambda functions to any cloud provider, or host it yourself.

## Next steps

Try it out in our [Getting started guide](./getting-started)
