exports.run = function(event, context) {
  // this will result in "[Object object]"
  context.fail({a: 1, b: false, c: { d: [1, 2] }});
}
