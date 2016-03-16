FROM iron/lambda-java8

ADD target/lambda-java-example-1.0-SNAPSHOT.jar ./lambda-java-example-1.0-SNAPSHOT.jar
RUN echo "{ \"firstName\":\"John\", \"lastName\":\"Doe\" }" >> ./payload.json
ENV PAYLOAD_FILE="payload.json"

CMD ["lambda-java-example-1.0-SNAPSHOT.jar", "example.Hello::myHandlerPOJO"]
