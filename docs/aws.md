Interacting with AWS Services
=============================

The node.js and Python stacks include SDKs to interact with other AWS services.
For Java you will need to include any such SDK in the JAR file.

## Credentials

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

    iron register <hub user>/s3-write

do this instead:

    iron register -e AWS_ACCESS_KEY_ID=<access key> \
                  -e AWS_SECRET_ACCESS_KEY=<secret key> \
                  <hub user>/s3-write

Alternatively, if you use `iron publish-function`, it will automatically
pick up the environment variables and forward them if valid ones are found.

```sh
export AWS_ACCESS_KEY_ID=<access key>
export AWS_SECRET_ACCESS_KEY=<secret key>
iron publish-function -function-name <hub user>/s3-write
```

If you have an existing image with the same name registered with IronWorker,
the environment variables will not simply be updated. You need to first delete
the code from HUD and then publish the function again. This will unfortunately
result in a new webhook URL for the function.

## Example: Using Lambda with IronWorker and Amazon Simple Notification Service

Lambda's premise of server-less computing requires a few infrastructural pieces
other than just the Docker image. First there needs to be a platform that can
run these Docker images on demand. Second, we need some way to invoke the
Lambda function based on an external event.

In this example, we will look at how to use IronWorker and Amazon Simple
Notification Service (SNS) to create a function that can search a given URL for
a user-specified keyword. You can build upon this, coupled with some storage
provider (like Amazon S3) to build a simple search engine.

The concepts introduced here can be used with any infrastructure that let's you
start a Docker container on some event. It is not tied to IronWorker.

The code for this example is located [here](../examples/sns/sns.js).

### Setup

Make sure you have an [IronWorker](https://www.iron.io/platform/ironworker/)
account. You can make one [here](https://www.iron.io/get-started/). You will
need a [Docker Hub](https://hub.docker.com) account to publish the Lambda function.

Also set up an [AWS
account](http://docs.aws.amazon.com/sns/latest/dg/SNSBeforeYouBegin.html) and [create a SNS topic](http://docs.aws.amazon.com/sns/latest/dg/CreateTopic.html). Call this topic `sns-example`. Carefully note the region the topic was created in. The region is found in the topic ARN.

You will also need credentials to use the AWS SDK. These credentials can be
obtained from the IAM and are in the form of an Access Key and Secret. See the
[AWS page](./aws.md) page for more information.

### Function outline

SNS can notify a variety of endpoints when a message is published to the topic.
One of these is an HTTPS URL. IronWorker provides an HTTPS URL that runs an
instance of the Docker image when the URL receives a POST.

In this example, we will manually publish messages to SNS, which will trigger
the webhook, our Lambda function will fetch the URL passed in the message,
search for the keyword in the response, and print out the count.

Here is the beginning of the function:

```js
var http = require('http');
var AWS = require('aws-sdk');
AWS.config.region = 'us-west-1';

function searchString(context, text, key) {
  // Global and Ignore case flags.
  var regex = new RegExp(key, 'gi');

  var results = [];
  var m;
  while ((m = regex.exec(text)) != null) {
    results.push(m);
  }

  console.log("Found", results.length, "instances of", key);
  context.succeed();
}

function searchBody(context, res, key) {
  if (res.statusCode === 200) {
    var body = "";
    res.on('data', function(chunk) { body += chunk.toString(); });
    res.on('end', function() { searchString(context, body, key); });
  } else {
    context.fail("Non-200 status code " + res.statusCode + " fetching '" + message.url + "'. Aborting.");
  }
}
```

Here we set up the various functions that will implement our Lambda function's
logic. Set the AWS region to the region in the SNS topic ARN, otherwise our
function will fail.

The `searchBody` function takes a node `http.ClientResponse` and gathers the
body data, then calls `searchString` to perform the regular expression match.
Finally each function invokes the `context.fail()` or `context.succeed()`
functions as appropriate. This is important, otherwise our function won't
terminate until it times out, even if execution was done.

#### Handling SNS event types.

SNS events send the payload as a JSON message to our webhook. These are passed
on to the Lambda function in the handler's `event` parameter. Each SNS message
contains a `Type` field. We are interested in two types - `Notification` and
`SubscriptionConfirmation`. The former is used to deliver published messages.

Before SNS can start sending messages to the subscriber, the subscriber has to
confirm the subscription. This is to prevent abuse. The
`SubscriptionConfirmation` type is used for this. Our function will have to
deal with both.

```js
exports.handler = function(event, context) {
    if (event.Type == 'Notification') {
      // ...
    }
    else if (event.Type == 'SubscriptionConfirmation') {
      // ...
    } else {
      console.log("unknown event.Type", event.Type);
      context.fail();
    }
};
```

We can use the SDK to confirm the subscription.

```js
var sns = new AWS.SNS();
var params = {
  Token: event.Token,
  TopicArn: event.TopicArn,
};
sns.confirmSubscription(params, function(err, data) {
  if (err) {
    console.log(err, err.stack);
    context.fail(err);
  } else {
    console.log("Confirmed subscription", data);
    console.log("Ready to process events.");
    context.done();
  }
});
```

The `Token` is unique and has to be sent to SNS to indicate that we are a valid
subscriber that the message was intended for. Once we confirm the subscription,
this run of the Lambda function is done and we can stop (`context.done()`).
SNS is now ready to run this Lambda function when we publish to the topic.

Finally we come to the event type we expect to receive most often --
`Notification`. In this case, we try to grab the url and keyword from the
message and run our earlier `searchBody()` function on it.

```js
try {
  var message = JSON.parse(event.Message);
  if (typeof message.url == "string" && typeof message.keyword == "string") {
    http.get(message.url, function(res) { searchBody(context, res, message.keyword); })
        .on('error', function(e) {
          context.fail(e);
        });
  } else {
    context.fail("Invalid message " + event.Message);
  }
} catch(e) {
  context.fail(e);
}
```

### Trying it out

With this function ready, we can Dockerize it and publish it to actually try it
out with SNS.

```sh
iron lambda create-function -function-name <Docker Hub username>/sns-example -runtime
nodejs -handler sns.handler sns.js
```

This will create a local docker image. The `publish-function` command will
upload this to Docker Hub and register it with IronWorker.

FIXME(nikhil): AWS credentials bit.

To be able to use the AWS SDK, you'll also need to set two environment
variables. The values must be your AWS credentials.

```sh
AWS_ACCESS_KEY_ID=<access key>
AWS_SECRET_ACCESS_KEY=<secret key>

iron publish-function -function-name <Docker Hub username>/sns-example:latest
```

Visit the published function's code page in the [IronWorker control
panel](https://hud.iron.io). You should see a cloaked field called "Webhook
URL". Copy this URL.

In the AWS SNS control panel, visit the `sns-example` topic. Click "Create
Subscription". Select the subscription type as HTTPS and paste the webhook URL.
Once you save this, the IronWorker task should have been launched and then
finished successfully with a "Confirmed subscription" message.

Now you can click the blue "Publish to topic" button on the AWS SNS control
panel. Select the message format as JSON and the contents as (for example):

```js
{
  "default": "{\"url\": \"http://www.econrates.com/reality/schul.html\", \"keyword\": \"blackbird\"}"
}
```

SNS will send the string in the `"default"` key to all subscribers. You should
be able to see that the IronWorker task has been run again. If everything went
well, it should have printed out a summary and exited successfully.

That's it, a simple, notification based, Lambda function.

## Example: Reading and writing to S3 Bucket

This example demonstrates modifying S3 buckets and using the included
ImageMagick tools in a node.js function. Our function will fetch an image
stored in a key specified by the event, resize it to a width of 1024px and save
it to another key.

The code for this example is located [here](../examples/s3/example.js).

The event will look like:

```js
{
    "bucket": "iron-lambda-demo-images",
    "srcKey": "waterfall.jpg", 
    "dstKey": "waterfall-1024.jpg"
}
```

The setup, imports and SDK initialization.

```js
var im = require('imagemagick');
var fs = require('fs');
var AWS = require('aws-sdk');

exports.run = function(event, context) {
  var bucketName = event['bucket']
  var srcImageKey = event['srcKey']
  var dstImageKey = event['dstKey']

  var s3 = new AWS.S3();
}
```

First we retrieve the source and write it to a local file so ImageMagick can
work with it.

```js
s3.getObject({
    Bucket: bucketName,
    Key: srcImageKey
  }, function (err, data) {

  if (err) throw err;

  var fileSrc = '/tmp/image-src.dat';
  var fileDst = '/tmp/image-dst.dat'
  fs.writeFileSync(fileSrc, data.Body)

});
```

The actual resizing involves using the identify function to get the current
size (we only resize if the image is wider than 1024px), then doing the actual
conversion to `fileDst`. Finally we upload to S3.

```js
im.identify(fileSrc, function(err, features) {
  resizeIfRequired(err, features, fileSrc, fileDst, function(err, resized) {
    if (err) throw err;
    if (resized) {
      s3.putObject({
        Bucket:bucketName,
        Key: dstImageKey,
        Body: fs.createReadStream(fileDst),
        ContentType: 'image/jpeg',
        ACL: 'public-read',
      }, function (err, data) {
        if (err) throw err;
        context.done()
      });
    } else {
      context.done();
    }
  });
});
```
