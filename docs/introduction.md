# Lambda everywhere.

AWS Lambda introduced server-less computing to the masses. Wouldn't it be nice
if you could run the same Lambda functions on any platform, in any cloud?
Iron.io is proud to release a set of tools that allow just this. Package your
Lambda function in a Docker container and run it anywhere with an environment
similar to AWS Lambda.

Using a job scheduler such as IronWorker, you can connect these functions to
webhooks and run them on-demand, at scale. You can also use a container
management system paired with a task queue to run these functions in
a self-contained, platform-independent manner.

## Use cases

Lambda functions are great for writing "worker" processes that perform some
simple, parallelizable task like image processing, ETL transformations,
asynchronous operations driven by Web APIs, or large batch processing.

All the benefits that containerization brings apply here. Our tools make it
easy to write containerized applications that will run anywhere without having
to fiddle with Docker and get the various runtimes set up. Instead you can just
write a simple function and have an "executable" ready to go.

TODO

## How does it work?

We provide base Docker images for the various runtimes that AWS Lambda
supports. The `ironcli` tool helps package up your Lambda function into
a Docker image layered on the base image. We provide a bootstrap script and
utilities that provide a AWS Lambda environment to your code. You can then run
the Docker image on any platform that supports Docker. This allows you to
easily move Lambda functions to any cloud provider, or host it yourself.

The Docker container has to be run with a certain configuration, described
[here](./docker-configuration.md)

## Next steps

Write and package your Lambda functions with our [Getting started
guide](./getting-started.md). [Here is the environment](./environment.md) that
Lambda provides. `ironcli lambda` lists the commands to work with Lambda
functions locally.

There is a short guide to [using Lambda with IronWorker](./ironworker.md).
Non-AWS Lambda functions do have the disadvantage of not having deep
integration with other AWS services. Much of the push-based actions can be
solved by redirecting the event through [SNS and using webhooks](./sns.md).
AWS APIs are of course available for use through the AWS SDK available to the
function. We explain how to deal with authentication in [this guide](./aws.md).

## Contributing

TODO
