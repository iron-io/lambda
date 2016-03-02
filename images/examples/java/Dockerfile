FROM iron/lambda-java

ADD target/lambda-java-example-1.0-SNAPSHOT.jar ./lambda-java-example-1.0-SNAPSHOT.jar

CMD ["lambda-java-example-1.0-SNAPSHOT.jar", "example.Hello::myHandlerPOJO", "{ \"firstName\":\"John\", \"lastName\":\"Doe\" }"]
