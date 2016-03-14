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

    AWS_ACCESS_KEY_ID
    AWS_SECRET_ACCESS_KEY
    IRON_WORKER_TOKEN
    IRON_WORKER_PROJECT_ID
    IRON_LAMBDA_TEST_IMAGE_PREFIX=<username to use for Docker images>

If you want to use staging, set:

    IRON_WORKER_HOST=staging-worker.iron.io

If you want email notifications on failures, set:
    SENDGRID_API_KEY=<key>

You can either set these in your shell or pass them to Docker.
Email support is disabled if the suite is not running on IronWorker. This is to
avoid noise while you are developing tests locally. If you want to test that
behaviour, comment out the `TASK_ID` check. Make sure to change the recipients
of the email!

The test suite running on IronWorker emails a set of IDs when a test fails.
See `main.go` `notifyFailure()` function for list of emails and from email (to
set up email filters).

### Running test suite

We use [Glide](http://glide.sh) for dependency management. Otherwise vendor
dependencies manually.

Run the Docker image, or:

    $ GOPATH/bin/glide install
    $ go build .
    $ ./test-suite

### Building Docker image

    make

Contributing
------------

### Deploying changes to test harness to IronWorker

(How do we prevent the harness from running tests when changes are being made?
Should we bother with this right now? Probably not.)

First update the local docker image following the instructions above. Then tag
the docker image

    docker tag irontest/test-suite irontest/test-suite:N

where N is the latest number that is not in use on [Docker
Hub](https://hub.docker.com/r/irontest/test-suite/tags/). You can also combine
the build and tag:

    docker build -t irontest/test-suite:N .

Push the new image:

    docker push irontest/test-suite:N

Register the image with Iron. You will need to pass various environment
variables for the tests to run properly. Please get these values from someone
in the company. The Iron Project is called Lambda Test Suite. The AWS
credentials are for user `lambdauser`.

    IRON_WORKER_PROJECT_ID=<project id> IRON_WORKER_TOKEN=<token> \
    $GOPATH/bin/ironcli register -e AWS_ACCESS_KEY_ID=<access key> \
                                 -e AWS_SECRET_ACCESS_KEY=<key> \
                                 -e IRON_WORKER_TOKEN=<token> \
                                 -e IRON_WORKER_PROJECT_ID=<project id> \
                                 -e IRON_LAMBDA_TEST_IMAGE_PREFIX=irontest \
                                 -e IRON_LAMBDA_TEST_LAMBDA_ROLE=<ARN for lambdauser> \
                                 -e SENDGRID_API_KEY=<key> \
                                 irontest/test-suite:N

Assuming you already have a dev environment setup with these variables, just
run:

    IRON_WORKER_PROJECT_ID=$IRON_WORKER_PROJECT_ID IRON_WORKER_TOKEN=$IRON_WORKER_TOKEN \
    $GOPATH/bin/ironcli register -e AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID \
                                 -e AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY \
                                 -e IRON_WORKER_TOKEN=$IRON_WORKER_TOKEN \
                                 -e IRON_WORKER_PROJECT_ID=$IRON_WORKER_PROJECT_ID \
                                 -e IRON_LAMBDA_TEST_IMAGE_PREFIX=irontest \
                                 -e IRON_LAMBDA_TEST_LAMBDA_ROLE=$IRON_LAMBDA_TEST_LAMBDA_ROLE \
                                 -e SENDGRID_API_KEY=$SENDGRID_API_KEY \
                                 irontest/test-suite:N

The test-suite will be scheduled to run periodically. If I (nikhil) understand
IronWorker correctly, the next run should automatically pick up the new image.

### The `lambda.test` file

This file describes some parameters of the test. It is a simple JSON file.

    {
      "handler": "test.run",
      "runtime": "nodejs",
      "name": "event",
      "event": {
        "key1": "value1",
      },
      "timeout" : 30
    }

Handler - The handler as defined by AWS lambda.

Runtime - AWS Lambda runtime name.

Name - Name of the test. The harness will create a Lambda function
`lambda-test-suite-<runtime>-<name>` based on this. The harness will also
create a docker image `lambda-test-suite-<runtime>-<name>` and a corresponding
IronWorker. If you change this, it is your responsibility to clean up existing
functions/images.

Event - JSON payload sent to the function and worker.

Timeout - the duration in seconds to wait for finishing AWS Lambda Function event processing.
If not specified the default value of 30 is used


### A note on Java tests

The Java build process is a little cumbersome. First, write your test. Keep it
in the lambdatest package. You'll want to use maven or similar to build the
package. The test directory layout will look like:

    test-suite/tests/java/test-simple/
    ├── lambda.test
    ├── pom.xml
    ├── src
    │   └── main
    │       └── java
    │           └── lambdatest
    │               └── TestFile.java
    ├── target
    │   ├── ...
    │   └── test-simple-1.0.jar
    └── test-build.jar

Due to the complications of uploading Java functions to AWS, the harness only
supports one style of Java functions, which is to create a JAR and upload the
JAR to AWS. This means, specifically for Java tests, the harness and tools will
exclude all files except the `test-build.jar`. You MUST copy the Maven
generated file from `target` to `test-build.jar` before using `add-test` or
other tools. The test harness does not support non-JSON payloads right now.
Logging is easiest to do by `System.out.println()`.

It seems like Java images don't always run from the test-suite. It could be
because they are large in size, or due to something else.

### Testing a test locally

`IRON_LAMBDA_TEST_LAMBDA_PREFIX=irontest` must be set.

The `local-image` tool will build a docker image out of a test directory. For
example:

    go run ./tools/local-image/main.go tests/node/test-context

Run it via:

    docker run --rm -it irontest/lambda-test-suite-nodejs-context

### Adding/Updating a test

The following environment variable must be set in addition to the ones above.

    IRON_LAMBDA_TEST_LAMBDA_ROLE=<fully qualified AWS role to execute lambda
    function as>

For execution role see [Getting Started][gs] and [Permissions Model][pm].

[gs]: http://docs.aws.amazon.com/lambda/latest/dg/get-started-create-function.html
[pm]: http://docs.aws.amazon.com/lambda/latest/dg/intro-permission-model.html

You MUST run this command every time you introduce a new test or make changes to an
existing test.

    go run ./tools/add-test/main.go tests/path/to/test/dir (e.g. tests/node/test-event)

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

Understand that the `test-suite` binary only runs tests for the directories it
can find in it's local `tests` directory. This means even after running the
`add-test` tool, the automated test harness will not run those tests. This
separation is good because you can now run the test locally as many times as
you like and fix any failures. To have the IronWorker deployed, automatic
test-harness pick up these new tests, you must recreate the `test-suite` Docker
image and publish+register it as specified at the beginning of this guide. Also
remember to add the tests to version control.

#### Java tests

The procedure for building Java tests is slightly involved. You will need Maven
in your path.

Copy one of the available tests to make changes:

    cp -R tests/java/test-resolution-ctx tests/java/test-new-test

You'll have to edit the following three files:

    # pom.xml
    Change artifactId to test name.

    # lambda.test
    Change test name and description and event.

    # Makefile
    Change the cp command's first argument.

The last step is because the local-image and add-test tools require a file
called `test-build.jar` for Java tests.

The Makefile provides a convenient script to build the test, copy the JAR and
register the test locally (for local testing).

### Removing a test

TODO
