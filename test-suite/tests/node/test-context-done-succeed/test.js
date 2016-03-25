var fs = require('fs');

exports.run = function(event, context) {
  fs.readFile("./test.js", context.done);
  console.log("done");
}
