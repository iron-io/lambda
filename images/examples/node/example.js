var uuid = require('uuid')

exports.run = function(event, context) {
  console.log("Payload is " + JSON.stringify(event))
  context.succeed("Newly minted uuid is " + uuid.v4());
}
