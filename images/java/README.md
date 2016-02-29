Support for running java Lambda functions.


Create an image with

    sudo docker build -t iron/lambda-java .


Running
-------

Compile your code and create jar file (compile example project in example folder with maven)

    mvn package

Run docker container inside folder with compiled jar file

    export handler="example.Hello::myHandlerPOJO"
    export payload='{ "firstName":"John", "lastName":"Doe" }'
    docker run -ti -e "handler=example.Hello::myHandler" -e "payload=asd" -e "JAR_ZIP_FILENAME=lambda-java-example.jar" -v $(pwd):/app -w /app iron/lambda-java

`handler` env var for package and handler function, `payload` env for payload (int, string, json), `JAR_ZIP_FILENAME` env for filename of compiled jar
