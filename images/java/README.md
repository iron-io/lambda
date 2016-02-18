Support for running java Lambda functions.


Create an image with ->  not working yet

    docker build -t iron/java-lambda .


Running
-------

Compile your code and create jar file (compile example project in example folder with maven)

    mvn package

Copy compiled jar file in current folder, then compile LambdaLauncher

    javac -cp ".:lambda-java-example-1.0-SNAPSHOT.jar:gson-2.6.1.jar" LambdaLaunchder.java

Create handler and payload env var

    export handler=example.Hello::myHandler
    export payload=123

Launch Lambda-Java

    java -cp ".:lambda-java-example-1.0-SNAPSHOT.jar:gson-2.6.1.jar" LambdaLaunchder

