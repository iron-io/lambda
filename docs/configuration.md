## Using the AWS SDK from Lambda functions.

Running Lambda functions outside of AWS means that we cannot automatically get
access to other AWS resources based on Lambda subsuming the execution role
specified with the function. Instead, when using the AWS APIs inside your
Lambda function (for example, to access S3 buckets), you will need to pass
these credentials explicitly.

### Using environment variables for the credentials

The easiest way to do this is to register the `AWS_ACCESS_KEY_ID` and
`AWS_SECRET_ACCESS_KEY` environment variables with IronWorker when registering
the Docker image. Then you may use the Environment Credentials loader provided
by the AWS SDK in various languages. Consider this node.js example:

    var AWS = require('aws-sdk');
    AWS.config.region = 'us-west-2';
    
    exports.run = function(event, context) {
      var s3bucket = new AWS.S3({
        params: {Bucket: event.bucket},
        credentials: new AWS.EnvironmentCredentials('AWS')
      });
      s3bucket.createBucket(function() {
        // Act on bucket here.
      });
    }

We pass the S3 object a credentials created from the environment variables.

Assuming you [packaged this function](./introduction.md) into a Docker image
`iron/s3-write` and pushed it to Docker Hub. Instead of just registering with
IronWorker as:

    ironcli register iron/s3-write

do this instead:

    ironcli register -e AWS_ACCESS_KEY_ID=<access key> \
                     -e AWS_SECRET_ACCESS_KEY=<secret key> \
                     iron/s3-write

Now, when you invoke the function, the AWS SDK will load the credentials from
the environment variables and your AWS API operations should work.
