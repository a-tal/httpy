# httpy example dockerfile.
# builds the example main application and the python worker module
FROM python:3.6 as build

ENV GOPATH="/go" \
    PATH="/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"

RUN apt-get update -qqy && \
    apt-get install -qqy pkg-config && \
    curl -sL#o go.tgz "https://dl.google.com/go/go1.10.2.linux-amd64.tar.gz" && \
    echo "4b677d698c65370afa33757b6954ade60347aaca310ea92a63ed717d7cb0c2ff *go.tgz" | sha256sum -c - && \
    tar -C /usr/local -xzf go.tgz && \
    rm go.tgz && \
    mkdir /go && \
    go version

WORKDIR /go/src/github.com/a-tal/httpy/example
ADD . /go/src/github.com/a-tal/httpy/

RUN go install -v ./...

FROM python:3.6
MAINTAINER Adam Talsma <adam@talsma.ca>

WORKDIR /src
COPY --from=build /go/bin/example /example
COPY example/*.py /src/

RUN pip install -q .

EXPOSE 8080
CMD /example
