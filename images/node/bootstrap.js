'use strict';

var fs = require('fs');

// Some notes on the semantics of the succeed(), fail() and done() methods.
// Tests are the source of truth!
// First call wins in terms of deciding the result of the function. BUT,
// subsequent calls also log. Further, code execution does not stop, even where
// for done(), the docs say that the "function terminates". It seems though
// that further cycles of the event loop do not run. For example:
// index.handler = function(event, context) {
//   context.fail("FAIL")
//   process.nextTick(function() {
//     console.log("This does not get logged")
//   })
//   console.log("This does get logged")
// }
// on the other hand:
// index.handler = function(event, context) {
//   process.nextTick(function() {
//     console.log("This also gets logged")
//     context.fail("FAIL")
//   })
//   console.log("This does get logged")
// }
//
// The same is true for context.succeed() and done() captures the semantics of
// both. It seems this is implemented simply by having process.nextTick() cause
// process.exit() or similar, because the following:
// exports.handler = function(event, context) {
//     process.nextTick(function() {console.log("This gets logged")})
//     process.nextTick(function() {console.log("This also gets logged")})
//     context.succeed("END")
//     process.nextTick(function() {console.log("This does not get logged")})
// };
//
// So the context object needs to have some sort of hidden boolean that is only
// flipped once, by the first call, and dictates the behavior on the next tick.
//
// In addition, the response behaviour depends on the invocation type. If we
// are to only support the async type, succeed() must return a 202 response
// code, not sure how to do this.
//
// Only the first 256kb, followed by a truncation message, should be logged.
//
// Also, the error log is always in a json literal
// { "errorMessage": "<message>" }
var Context = function() {
  var concluded = false;

  var contextSelf = this;

  // The succeed, fail and done functions are public, but access a private
  // member (concluded). Hence this ugly nested definition.
  this.succeed = function(result) {
    if (concluded) {
      return
    }

    // We have to process the result before we can conclude, because otherwise
    // we have to fail. This means NO EARLY RETURNS from this function without
    // review!
    if (result === undefined) {
      result = null
    }

    var str;
    try {
      str = JSON.stringify(result)
      // Succeed does not output to log, it only responds to the HTTP request.
    } catch(e) {
      // Set X-Amz-Function-Error: Unhandled header
      return contextSelf.fail("Unable to stringify body as json: " + e);
    }

    // FIXME(nikhil): Return 202 or 200 based on invocation type and set response
    // to result. Should probably be handled externally by the runner/swapi.

    // OK, everything good.
    concluded = true;
    process.nextTick(function() { process.exit(0) })
  }

  this.fail = function(error) {
    if (concluded) {
      return
    }

    concluded = true
    process.nextTick(function() { process.exit(1) })

    if (error === undefined) {
      error = null
    }

    // FIXME(nikhil): Truncated log of error, plus non-truncated response body
    var errstr = "fail() called with argument but a problem was encountered while converting it to a to string";

    // The semantics of fail() are weird. If the error is something that can be
    // converted to a string, the log output wraps the string in a JSON literal
    // with key "errorMessage". If toString() fails, then the output is only
    // the error string.
    try {
      if (error === null) {
        errstr = null
      } else {
        errstr = error.toString()
      }
      console.log(JSON.stringify({"errorMessage": errstr }))
    } catch(e) {
      // Set X-Amz-Function-Error: Unhandled header
      console.log(errstr)
    }
  }

  this.done = function() {
    var error = arguments[0];
    var result = arguments[1];
    if (error) {
      contextSelf.fail(error)
    } else {
      contextSelf.succeed(result)
    }
  }
}

Context.prototype.getRemainingTimeInMillis = function() {
  // FIXME(nikhil): Obviously need to feed this in to the program.
  return 42;
}

var getEnv = function(name) {
  return process.env[name] || "";
}

var makeCtx = function() {
  var ctx = new Context();
  Object.defineProperties(ctx, {
    "functionName": {
      get: function() {
        return getEnv("AWS_LAMBDA_FUNCTION_NAME");
      },
      enumerable: true,
    },
    "functionVersion": {
      value: "$LATEST",
      enumerable: true,
    },
    "invokedFunctionArn": {
      // FIXME(nikhil): Should be filled in.
      value: "",
      enumerable: true,
    },
    "memoryLimitInMB": {
      // FIXME(nikhil): Should be filled in.
      value: 256,
      enumerable: true,
    },
    "awsRequestId": {
      get: function() {
        return getEnv("TASK_ID");
      },
      enumerable: true,
    },
    "logGroupName": {
      // FIXME(nikhil): Should be filled in.
      value: "",
      enumerable: true,
    },
    "logStreamName": {
      // FIXME(nikhil): Should be filled in.
      value: "",
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
    var entry = parts[1];
    try {
      var mod = require('./'+script);
      if (mod[entry] === undefined) {
        throw "Handler '" + entry + "' missing on module '" + script + "'"
      }

      if (typeof mod[entry] !== 'function') {
        throw "TypeError: " + (typeof mod[entry]) + " is not a function"
      }

      mod[entry](payload, makeCtx())
    } catch(e) {
      if (typeof e === 'string') {
        console.log(e)
      } else {
        console.log(e.message)
      }
    }
  } else {
    console.error("bootstrap: No script specified")
  }
}

run()
