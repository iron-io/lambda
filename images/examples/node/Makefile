IMAGE=iron/lambda-node-example

image: Dockerfile example.js node_modules
	docker build -t $(IMAGE) .

run:
	docker run --rm -it $(IMAGE)
