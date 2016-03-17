IMAGE=iron/lambda-java8

image: Dockerfile src/main/java/io/iron/lambda/*.java
	mvn package
	docker build -t $(IMAGE) .
