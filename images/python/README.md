Support for running python Lambda functions.

Create an image with

    docker build -t iron/lambda-python .

This sets up a python stack and installs some deps to provide the Lambda runtime.

Running
-------

Does not support payload (AWS Lambda 'event') yet. 
Expects the lambda files places inside or mounted to docker image.
Two environment variables should be provided:
* HANDLER is the name of python function to call in format <python-module-name>.<top-level-function-name>
* PAYLOAD_FILE is the location of the payload file in json format to pass into function in the event variable

To locate the specified python module the /mnt folder is seached first and the default python module import algorithm on failback

Example: 

for the fancyFunction inside fancy.py
    // fancy.py 
    def fancyFunction(event, context):

the HANDLER should be fancy.fancyFunction

To run the image with payload.json and fancy.py is the working directory use the following command
    docker run --rm -it -v `pwd`:/mnt -e HANDLER=fancy.fancyFunction -e PAYLOAD_FILE=/mnt/payload.json iron/lambda-python

In Lambda you'd submit the HANDLER parameter in the call to `create-function`.
