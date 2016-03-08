dockerize:
	$(GOPATH)/bin/glide install
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build .
	docker build -t irontest/test-suite .
	rm -f ./test-suite
	@echo "\nRemember to tag before deploying!"
