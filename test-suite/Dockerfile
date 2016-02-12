# Dockerfile to create a Docker image to run this on IronWorker
# First cross compile lambda-test-suite for Linux
FROM iron/go
ADD ./lambda-test-suite ./app/lambda-test-suite
CMD ["/app/lambda-test-suite"]
