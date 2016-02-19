exports.run = function(event, context) {
  // context.succeed.
  console.log("context.succeed is a function",
              typeof context.succeed == 'function');
  console.log("context.succeed takes one argument", context.succeed.length === 1);

  // context.fail.
  console.log("context.fail is a function", typeof context.fail == 'function');
  console.log("context.fail takes one argument", context.fail.length === 1);

  // context.done.
  // Although done accepts 2 arguments (both optional), the length value is
  // 0 on Lambda.
  console.log("context.done is a function", typeof context.done == 'function');
  console.log("context.done takes zero arguments", context.done.length === 0);

  // context.getRemainingTimeInMillis.
  console.log("context.getRemainingTimeInMillis is a function",
              typeof context.getRemainingTimeInMillis == 'function')
  console.log("context.getRemainingTimeInMillis takes zero arguments",
              context.getRemainingTimeInMillis.length === 0)

}
