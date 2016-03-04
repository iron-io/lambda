exports.run = function(event, context) {
  console.log("This gets logged.");
  process.nextTick(function() { console.log("This gets logged too."); })
  context.fail("This also gets logged.")
  process.nextTick(function() { console.log("This is not logged."); })
}
