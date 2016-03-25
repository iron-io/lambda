var assert = require('assert');
exports.run = function(event, context) {
  // context.succeed.
  assert.ok(typeof context.succeed == 'function',
              "context.succeed is a function");
  assert.ok(context.succeed.length === 1, "context.succeed takes one argument");

  // context.fail.
  assert.ok(typeof context.fail == 'function', "context.fail is a function");
  assert.ok(context.fail.length === 1, "context.fail takes one argument");

  // context.done.
  // Although done accepts 2 arguments (both optional), the length value is
  // 0 on Lambda.
  assert.ok(typeof context.done == 'function', "context.done is a function");
  assert.ok(context.done.length === 0, "context.done takes zero arguments");

  // context.getRemainingTimeInMillis.
  assert.ok(typeof context.getRemainingTimeInMillis == 'function',
            "context.getRemainingTimeInMillis is a function");
  assert.ok(context.getRemainingTimeInMillis.length === 0,
            "context.getRemainingTimeInMillis takes zero arguments");

  var timeLeft = context.getRemainingTimeInMillis();
  assert.ok(timeLeft >= 0,
            "context.getRemainingTimeInMillis returns a non-negative number");
  setTimeout(function() {
    var newLeft = context.getRemainingTimeInMillis();
    assert.ok(newLeft < timeLeft, "subsequent call to context.getRemainingTimeInMillis() should report lower value.");
  }, 500);

  // We can't check for equality since the AWS and Iron naming convention is different.
  assert.ok(typeof context.functionName == 'string' && context.functionName.match('context'), "context.functionName contains 'context'.");

  // We only support latest for now.
  assert.ok(typeof context.functionVersion == 'string' && context.functionVersion == "$LATEST", "context.functionVersion is $LATEST");

  assert.ok(typeof context.awsRequestId == 'string' && context.awsRequestId.length > 0, "context.awsRequestId is defined.");

  assert.ok(typeof context.memoryLimitInMB == 'string' && parseInt(context.memoryLimitInMB) >= 0, "context.memoryLimitInMB is a string representing a non-negative number.");

  context.done('ok')
}
