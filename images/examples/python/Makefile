IMAGE=iron/lambda-python-example

image: Dockerfile 
	docker build -t $(IMAGE) .

run:
	docker run --rm -it -e PAYLOAD_FILE=/mnt/example-payload.json -v `pwd`:/mnt $(IMAGE)
