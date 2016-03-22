FROM iron/lambda-python2.7

RUN apk --no-cache add imagemagick py-setuptools
RUN easy_install-2.7 Wand

ADD example.py ./example.py

CMD ["example.run"]
