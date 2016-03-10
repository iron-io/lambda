Support for running python Lambda functions.

Create an image with
```
    make image
```
This sets up a python stack and installs some deps to provide the Lambda runtime.

Running
-------

Expects the lambda files places inside or mounted to docker image.
The following mandatory parameters should be provided:
* The command line argument with the name of python function to call in format `<python-module-name>.<top-level-function-name>`
* The environment variable `PAYLOAD_FILE` with the location of the payload file in json format to pass into function in the event variable

To locate the specified python module the /mnt folder is seached first and the default python module import algorithm on failback

Example:

for the fancyFunction inside `fancy.py`
```
    // fancy.py
    def fancyFunction(event, context):
        print (event)
```
the command line argument should be `fancy.fancyFunction`

To run the image with `payload.json` and `fancy.py` in the working directory use the following command
```
    docker run --rm -it -v `pwd`:/mnt -e PAYLOAD_FILE=/mnt/payload.json iron/lambda-python fancy.fancyFunction
```
