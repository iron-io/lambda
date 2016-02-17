# Dockerfile to create a Docker image to run this on IronWorker
# First cross compile lambda-test-suite for Linux
FROM iron/go
WORKDIR /app
ADD ./test-suite ./test-suite
ADD ./tests ./tests
CMD ["/app/test-suite"]
