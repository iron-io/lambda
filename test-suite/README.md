This is a set of tests that run various lambda functions on AWS and Iron and
check if the output is the same. The intention is to offer compliance with the
AWS Lambda API to the best of our abilities. These tests should run in
a scheduled worker and failure should lead to emails/logging.

The test-suite is a collection of lambda functions and description
files. We'll read the description file and run a Lambda function. Then
do the same for iron.

The test suite proceeds like this:

For each test in suite:
- Invoke function on Lambda
- Invoke function on Iron in parallel. Both must be async invocations.
- We need to log the output somewhere. For Lambda this is likely cloudwatch.
  For Iron this is the program output. Then we need to compare.

### Configuration

The following environment variables must be set:

    AWS_ACCESS_KEY
    AWS_SECRET_KEY
    IRON_WORKER_TOKEN
    IRON_WORKER_PROJECT_ID
    IRON_LAMBDA_TEST_IMAGE_PREFIX=<username to use for Docker images>

If you want to use staging, set:

    IRON_WORKER_HOST=staging-worker.iron.io

You can either set these in your shell or pass them to Docker.

### Running test suite

Run the Docker image, or:

    $ go build .
    $ ./lambda-test-suite

### Building Docker image

    $ GOOS=linux GOARCH=amd64 go build .
    $ docker build -t iron/lambda-test-suite .

Contributing
------------

### Deploying changes to test harness to IronWorker

NOTE: This is required when you change how the test harness program
`lambda-test-suite` works. If you only change a test, see `Updating a test`
below.

How do we prevent the harness from running tests when changes are being made?
Should we bother with this right now? Probably not.

### The `lambda.test` file

This file describes some parameters of the test. It is a simple JSON file.

    {
      "handler": "test.run",
      "runtime": "nodejs",
      "name": "event",
      "event": {
        "key1": "value1",
      }
    }

Handler - The handler as defined by AWS lambda.

Runtime - AWS Lambda runtime name.

Name - Name of the test. The harness will create a Lambda function
`lambda-test-suite-<runtime>-<name>` based on this. The harness will also
create a docker image `lambda-test-suite-<runtime>-<name>` and a corresponding
IronWorker. If you change this, it is your responsibility to clean up existing
functions/images.

Event - JSON payload sent to the function and worker.

### Adding/Updating a test

The following environment variable must be set in addition to the ones above.

    IRON_LAMBDA_TEST_LAMBDA_ROLE=<fully qualified AWS role to execute lambda
    function as>

For execution role see [Getting Started][gs] and [Permissions Model][pm].

[gs]: http://docs.aws.amazon.com/lambda/latest/dg/get-started-create-function.html
[pm]: http://docs.aws.amazon.com/lambda/latest/dg/intro-permission-model.html

You MUST run this command every time you introduce a new test or make changes to an
existing test.

    go run ./tools/add-test.go tests/path/to/test/dir (e.g. tests/node/test-event)

Adding a test does the following:

1. Lambda: Zips any files/directories in the test dir (except `lambda.test`) and
   creates/updates the AWS Lambda function.
1. Iron:
  1. Generates a new UUID. This UUID will be used as the tag for the docker
     image to identify it distinctly from older instances.
  2. Creates a new Docker image with name derived from the lambda.test file,
     and tagged with UUID.
  3. Publishes this image to Hub.
  4. Register's image with UUID with IronWorker, replacing the older
     association. This means the Worker with name derived from `lambda.test`
     always runs the latest UUID.

### Removing a test

TODO
