Support for running java Lambda functions.

Compile lambda launcher and create an lambda-java base image:

    mvn package
    sudo docker build -t iron/lambda-java .


Running
-------

Compile your code with maven (you can compile example project in examples folder)

    mvn package

Then build docker image and run docker container, you can set handler and payload in Dockerfile (CMD string, second param for handler, third for payload)

    docker build -t lambda-java-my .
    docker run lambda-java-my

Or you can set handler and payload in env vars

    docker run -ti -e "handler=example.Hello::myHandlerIO" -e "payload=test_string" lambda-java-my

Examples of `handler` and `payload` params:

    handler=example.Hello::myHandlerInt
    payload=1122

    handler=example.Hello::myHandlerString
    payload=test_string

    handler=example.Hello::myHandlerIO
    payload=test_input_line

    handler=example.Hello::myHandlerMap
    payload={\"zero\":\"Zero Element\",\"third\":\"Third Element!\"}

    handler=example.Hello::myHandlerList
    payload=[{\"user\":\"1\",\"pass\":\"2\",\"secretCode\":\"3\"}]

    handler=example.Hello::myHandlerPOJO
    payload={ \"firstName\":\"John\", \"lastName\":\"Doe\" }
