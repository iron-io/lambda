var im = require('imagemagick');
var http = require('http');
var fs = require('fs');

var download = function(url, fn, cb) {
  http.get(url, function(resp) {
    resp.pipe(fs.createWriteStream(fn)).on('close', cb);
  });
}

exports.run = function(event, context) {
  download(event.url, event.fn, function(err) {
    if (err) throw err;
    im.readMetadata(event.fn, function(err, metadata){
      if (err) throw err;
      console.log(metadata.exif.model);
      context.succeed();
    })
  })
}

// For running without Lambda support
// exports.run({url: 'http://www.exiv2.org/include/img_1771.jpg', fn: 'image.jpeg'}, {})
