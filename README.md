This is a set of tools, tests and libraries to convert Amazon AWS Lambda
functions into Docker images that can run on any computer or cloud provider.

Java, Python and node.js are supported to various levels of compatibility.

Instructions on using this repository to convert Lambda functions to Docker
images are [in the introduction](https://github.com/iron-io/lambda/blob/master/docs/introduction.md).

ironcli lambda support is in the [ironcli][ironcli] repository.

[ironcli]: https://github.com/iron-io/ironcli

## ./lambda

Library to Dockerize lambda functions. This is used by the test-suite and
[ironcli][ironcli].

## ./images

Lambda compatible docker base images, bootstrap code and examples of creating
custom images.

## ./test-suite

Harness and tests to run Lambda functions on AWS and IronWorker and ensure
compatibility. See the test-suite README for more information.

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md).

Based on [previous work](https://github.com/vlopatkin/iron-lambda) by Vitaly Lopatkin.
