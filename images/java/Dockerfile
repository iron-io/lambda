FROM iron/java:1.8

WORKDIR /LambdaLauncher

COPY target/lambda-* /LambdaLauncher/lambda.jar
COPY LambdaLauncher.sh /LambdaLauncher/

ENTRYPOINT ["./LambdaLauncher.sh"]
