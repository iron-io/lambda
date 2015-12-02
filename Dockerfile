FROM iron/base

ADD ironcli /usr/local/bin/iron

ENTRYPOINT ["/usr/local/bin/iron"]
