'use strict';

var fs = require('fs');

var Context = function() {}

Context.prototype.succeed = function(result) {
  if (!result) {
    return
  }

  var str;
  try {
    str = JSON.stringify(result)
  } catch(e) {
    // Set X-Amz-Function-Error: Unhandled header
  }

  // FIXME(nikhil): Return 202 or 200 based on invocation type and set response
  // to result. Should probably be handled externally by the runner/swapi.
}

Context.prototype.fail = function(error) {
  if (error) {
    try {
      var str = JSON.stringify(error);
      // FIXME(nikhil): Truncated log of str, plus non-truncated response body
      console.error(str)
    } catch(e) {
      // Set X-Amz-Function-Error: Unhandled header
    }
  }

  process.exit(1)
}

Context.prototype.done = function(error, result) {
  if (error) {
    this.fail(error)
  } else {
    this.succeed(result)
  }
}

Context.prototype.getRemainingTimeInMillis = function() {
  // FIXME(nikhil): Obviously need to feed this in to the program.
  return 42;
}

var makeCtx = function() {
  var ctx = new Context();
  Object.defineProperties(ctx, {
    "functionName": {
      // FIXME(nikhil): Should be filled in.
      value: "hello",
      enumerable: true,
    },
    "functionVersion": {
      // FIXME(nikhil): Should be filled in.
      value: "hello",
      enumerable: true,
    },
    "invokedFunctionArn": {
      // FIXME(nikhil): Should be filled in.
      value: "hello",
      enumerable: true,
    },
    "memoryLimitInMB": {
      // FIXME(nikhil): Should be filled in.
      value: 256,
      enumerable: true,
    },
    "awsRequestId": {
      // FIXME(nikhil): Should be filled in.
      value: "hello",
      enumerable: true,
    },
    "logGroupName": {
      // FIXME(nikhil): Should be filled in.
      value: "iron",
      enumerable: true,
    },
    "logStreamName": {
      // FIXME(nikhil): Should be filled in.
      value: "your-worker",
      enumerable: true,
    },
    "identity": {
      // FIXME(nikhil): Should be filled in.
      value: null,
      enumerable: true,
    },
    "clientContext": {
      // FIXME(nikhil): Should be filled in.
      value: null,
      enumerable: true,
    },
  });

  return ctx;
}

function run() {
  // FIXME(nikhil): Check for file existence and allow non-payload.
  var payload = {};
  var path = process.env["PAYLOAD_FILE"];
  if (path) {
    try {
      var contents = fs.readFileSync(path);
      payload = JSON.parse(contents);
    } catch(e) {
      console.error("bootstrap: Error reading payload file", e)
    }
  }

  if (process.argv.length > 2) {
    var handler = process.argv[2];
    var parts = handler.split('.');
    // FIXME(nikhil): Error checking.
    var script = parts[0];
    try {
      var mod = require('./'+script);
      mod[parts[1]](payload, makeCtx())
    } catch(e) {
      console.error("bootstrap: Error running handler", e)
    }
  } else {
    console.error("bootstrap: No script specified")
  }
}

run()
