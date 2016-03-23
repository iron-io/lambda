var i = require('./this is an import');
exports.run = function(event, context) {
  console.log("start")
  context.succeed(i.answer);
}
