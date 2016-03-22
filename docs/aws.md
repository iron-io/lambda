## Using the AWS SDK from Lambda functions.

Running Lambda functions outside of AWS means that we cannot automatically get
access to other AWS resources based on Lambda subsuming the execution role
specified with the function. Instead, when using the AWS APIs inside your
Lambda function (for example, to access S3 buckets), you will need to pass
these credentials explicitly.

### Using environment variables for the credentials

The easiest way to do this is to pass the `AWS_ACCESS_KEY_ID` and
`AWS_SECRET_ACCESS_KEY` environment variables to Docker when you run the image.

```sh
docker run -e AWS_ACCESS_KEY_ID=<access key> -e AWS_SECRET_ACCESS_KEY=<secret key> <image>
```

The various AWS SDKs will automatically pick these up.

    var AWS = require('aws-sdk');
    AWS.config.region = 'us-west-2';

    exports.run = function(event, context) {
      var s3bucket = new AWS.S3({
        params: {Bucket: event.bucket},
      });
      s3bucket.createBucket(function() {
        // Act on bucket here.
      });
    }

### Credentials on IronWorker

Assuming you [packaged this function](./introduction.md) into a Docker image
`iron/s3-write` and pushed it to Docker Hub. Instead of just registering with
IronWorker as:

    ironcli register <hub user>/s3-write

do this instead:

    ironcli register -e AWS_ACCESS_KEY_ID=<access key> \
                     -e AWS_SECRET_ACCESS_KEY=<secret key> \
                     <hub user>/s3-write

Alternatively, if you use `ironcli publish-function`, it will automatically
pick up the environment variables and forward them if valid ones are found.

```sh
export AWS_ACCESS_KEY_ID=<access key>
export AWS_SECRET_ACCESS_KEY=<secret key>
ironcli publish-function -function-name <hub user>/s3-write
```

If you have an existing image with the same name registered with IronWorker,
the environment variables will not simply be updated. You need to first delete
the code from HUD and then publish the function again. This will unfortunately
result in a new webhook URL for the function.
