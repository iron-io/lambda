# Environment

The base images strive to provide the same environment that AWS provides to
Lambda functions. This page describes it and any incompatibilities between AWS
Lambda and Dockerized Lambda.

## nodejs

* node.js version [0.10.42][nodev]. Thanks to Michael Hart for creating the
  smaller, Alpine Linux based image.
* ImageMagick version [6.9.3][magickv] and nodejs [wrapper 6.9.3][magickwrapperv]
* aws-sdk version [2.2.12][awsnodev]

[nodev]: https://github.com/mhart/alpine-node/blob/f025a0516b87e2a505c6be4ff2c7bf485a95dc5a/Dockerfile
[magickv]: https://pkgs.alpinelinux.org/package/main/x86_64/imagemagick
[magickwrapperv]: https://www.npmjs.com/package/imagemagick
[awsnodev]: https://aws.amazon.com/sdk-for-node-js/

### Context object

TODO

## Python 2.7

* CPython [2.7.11][pythonv]
* boto3 (Python AWS SDK) [1.2.3][botov].

[pythonv]: https://hub.docker.com/r/iron/python/tags/
[botov]: https://github.com/boto/boto3/releases/tag/1.2.3

### Context object

TODO

## Java 8

* OpenJDK Java Runtime [1.8.0][javav]

[javav]: https://hub.docker.com/r/iron/java/tags/

### Context object

TODO
