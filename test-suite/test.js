var st = require('stack-trace');

function b() {
  throw new Error("YO");
}

function a() {
  b();
}

try {
  a();
} catch(e) {
  console.log(st.parse(e))
}
