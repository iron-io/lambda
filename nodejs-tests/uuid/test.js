var uuid = require('uuid')
exports.run = function(event, context) {
	context.succeed(uuid.v4())
}
