# AWS provides node 0.10.36. Since node follows semver, we can safely use any 0.10.x.
FROM mhart/alpine-node:0.10

WORKDIR /app

# Install ImageMagick and AWS SDK as provided by Lambda.
RUN apk update
RUN apk add imagemagick
RUN npm install aws-sdk@2.2.32 imagemagick
RUN rm -rf /var/cache/apk/*
RUN npm cache clear

# ironcli should forbid this name
ADD bootstrap.js /app/lambda-bootstrap.js

# Run the handler, with a payload in the future.

ENTRYPOINT ["node", "./lambda-bootstrap"]
