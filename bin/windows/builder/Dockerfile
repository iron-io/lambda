FROM monokrome/wine
MAINTAINER Peter Yeremenko <peter@iron.io>

RUN apt-get update -y
RUN apt-get install -y \
		build-essential \
		gcc \
		libc6-dev \
		make \
		imagemagick \
		bzr \
		git \
		mercurial \
		subversion \
		openssh-client \
		curl \
		wget \ 
		ca-certificates

ENV GOLANG_VERSION 1.3.3

RUN curl -sSL https://golang.org/dl/go$GOLANG_VERSION.src.tar.gz \
		| tar -v -C /usr/src -xz

RUN cd /usr/src/go/src && ./make.bash --no-clean 2>&1

ENV PATH /usr/src/go/bin:$PATH

RUN mkdir -p /go/src /go/bin && chmod -R 777 /go
ENV GOPATH /go
ENV PATH /go/bin:$PATH
WORKDIR /go

COPY go-wrapper /usr/local/bin/

ENV GOLANG_CROSSPLATFORMS \
	darwin/amd64 \
	linux/amd64 \
	windows/amd64

ENV GOARM 5

RUN cd /usr/src/go/src \
	&& set -ex \
	&& for platform in $GOLANG_CROSSPLATFORMS; do \
		GOOS=${platform%/*} \
		GOARCH=${platform##*/} \
		./make.bash --no-clean 2>&1; \
	done

COPY build.sh /build.sh

RUN mkdir /home/wix
COPY bin.tar.gz /home/wix/
RUN tar -zxf /home/wix/bin.tar.gz -C /home/wix/

ENTRYPOINT ["/bin/sh", "/build.sh"]