IMAGE=iron/lambda-python-aws-example

image: Dockerfile 
	docker build -t $(IMAGE) .

run:
	docker run --rm -it -e PAYLOAD_FILE=/mnt/example-payload.json -e AWS_ACCESS_KEY_ID=change-here -e AWS_SECRET_ACCESS_KEY=change-here -v `pwd`:/mnt $(IMAGE)
