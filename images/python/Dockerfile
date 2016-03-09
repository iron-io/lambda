FROM iron/python:2.7

WORKDIR /app

RUN apk update && apk add py-pip && rm -rf /var/cache/apk/* && pip install --upgrade pip --no-cache-dir 
RUN pip install -I boto3==1.2.3 --no-cache-dir 

# ironcli should forbid this name
ADD bootstrap.py /app/lambda-bootstrap.py

ENTRYPOINT ["python", "./lambda-bootstrap.py"]
