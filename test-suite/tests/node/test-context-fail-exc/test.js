exports.run = function(event, context) {
  // Cannot use Error() until we comply with AWS stack trace.
  var badobj = { toString: function() { throw "FAIL"; } };
  context.fail(badobj);
}
