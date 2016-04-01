Import existing AWS Lambda functions
====================================

The [ironcli](https://github.com/iron-io/ironcli/) tool includes a set of
commands to act on Lambda functions. Most of these are described in
[getting-started](./getting-started.md). One more subcommand is `aws-import`.

If you have an existing AWS Lambda function, you can use this command to
automatically convert it to a Docker image that is ready to be deployed on
other platforms.

### Credentials

To use this, either have your AWS access key and secret key set in config
files, or in environment variables. In addition, you'll want to set a default
region. You can use the `aws` tool to set this up. Full instructions are in the
[AWS documentation][awscli].

[awscli]: http://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html#cli-config-files

### Importing

Assuming you have a lambda function named `my-function`, the following command:

```sh
ironcli lambda aws-import my-function
```

will import the function code to a directory called `./my-function`. It will
then create a docker image called `my-function`.

If you only want to download the code, pass the `-download-only` flag. The
`-region` and `-profile` flags are available similar to the `aws` tool to help
you tweak the settings on a command level. If you want to call the docker image
something other than `my-function`, pass the `-image <new name>` flag. Finally,
you can import a different version of your lambda function than the latest one
by passing `-version <version>.`
