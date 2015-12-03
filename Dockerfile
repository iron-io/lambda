FROM iron/busybox

ADD ironcli /usr/local/bin/iron

ENTRYPOINT ["/usr/local/bin/iron"]
