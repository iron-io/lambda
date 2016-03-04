exports.run = function(event, context) {
  // Cannot use Error() until we comply with AWS stack trace.
  var badobj = { toJSON: function() { throw "FAIL"; } };
  // This fails.
  context.succeed(badobj);
}
