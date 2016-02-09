var i = require('./this is an import');
exports.run = function(event, context) {
  context.succeed(i.answer);
}
