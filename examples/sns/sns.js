// If you make changes to this file, make sure the SNS example documentation is
// updated.
//
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
          
exports.handler = function(event, context) {
    if (event.Type == 'Notification') {
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
    }
    else if (event.Type == 'SubscriptionConfirmation') {
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
    } else {
      console.log("unknown event.Type", event.Type);
      context.fail();
    }
};
