# Go Dockerfile

# Base alpine golang image
FROM golang:1.8.0-alpine

MAINTAINER Jarvis Prestidge "jarvisprestidge@gmail.com"

# Add all source code
ADD . /go/src/github.com/jarvisprestidge/sitemap

# Move to repo
WORKDIR /go/src/github.com/jarvisprestidge/sitemap

RUN apk add --no-cache git \
	&& go get golang.org/x/net/html \
    && go get github.com/onsi/ginkgo/ginkgo \
    && go get github.com/onsi/gomega \
    && apk del git

# Install the binary
RUN go install github.com/jarvisprestidge/sitemap

# Indicate the binary as our entrypoint
ENTRYPOINT "/bin/ash"
