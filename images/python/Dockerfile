FROM iron/python:2.7

WORKDIR /app

RUN \
  apk --no-cache add py-setuptools \
\
 && apk --no-cache add db \
 && apk --no-cache add --virtual build-deps python-dev musl-dev db-dev gcc \
 && easy_install-2.7 "bsddb3==5.3.0" \
 && apk del build-deps \
\
 && apk --no-cache add py-curl \
 && easy_install-2.7 "urlgrabber==3.9.1" \
\
 && easy_install-2.7 "botocore==1.4.1" \
 && easy_install-2.7 "boto3==1.2.6" \
\
 && rm  -rf /tmp/*

# ironcli should forbid this name
ADD bootstrap.py /app/lambda-bootstrap.py

ENTRYPOINT ["python", "./lambda-bootstrap.py"]
