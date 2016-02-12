exports.run = function(event, context) {
  keys = Object.keys(event);
  keys.sort();
  for (var i = 0; i < keys.length; i++) {
    console.info(keys[i], event[keys[i]])
  }
  context.succeed();
}
