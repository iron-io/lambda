Support for running java Lambda functions.

Compile lambda launcher and create an lambda-java base image:

    make

Running
-------

Compile your code with maven (you can compile example project in examples folder)

    mvn package

Then build docker image and run docker container, you can set handler and payload in Dockerfile (CMD string, second param for handler, third for payload)

    docker build -t lambda-java-my .
    docker run lambda-java-my

Or you can set handler and payload in env vars

    docker run -ti -e "HANDLER=example.Hello::myHandlerIO" -e "PAYLOAD_FILE=test_string" lambda-java-my

Examples of `handler` and `payload` params:

    HANDLER=example.Hello::myHandlerInt
    PAYLOAD_FILE=1122

    HANDLER=example.Hello::myHandlerString
    PAYLOAD_FILE=test_string

    HANDLER=example.Hello::myHandlerIO
    payload=test_input_line

    HANDLER=example.Hello::myHandlerMap
    PAYLOAD_FILE={\"zero\":\"Zero Element\",\"third\":\"Third Element!\"}

    HANDLER=example.Hello::myHandlerList
    PAYLOAD_FILE=[{\"user\":\"1\",\"pass\":\"2\",\"secretCode\":\"3\"}]

    HANDLER=example.Hello::myHandlerPOJO
    PAYLOAD_FILE={ \"firstName\":\"John\", \"lastName\":\"Doe\" }
