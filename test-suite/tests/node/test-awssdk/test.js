var AWS = require('aws-sdk');
AWS.config.region = 'us-west-2';

exports.run = function(event, context) {
  var s3bucket = new AWS.S3({params: {Bucket: event.bucket}, credentials: new AWS.EnvironmentCredentials('AWS')});
  s3bucket.createBucket(function() {
    var params = {Key: 'myKey', Body: 'Hello!'};
    s3bucket.upload(params, function(err, data) {
      if (err) {
        console.log("Error uploading data: ", err);
      } else {
        console.log("Successfully uploaded data to", params.Key, params.Body);
      }
      context.succeed();
    });
  });
}
