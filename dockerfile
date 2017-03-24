#
# Go Dockerfile
#

# Base alpine golang image
FROM golang:1.8.0-alpine

MAINTAINER Jarvis Prestidge "jarvisprestidge@gmail.com"

# Make the source code path
RUN mkdir -p /go/src/github.com/jarvisprestidge/domain-web-crawler

# Add all source code
ADD . /go/src/github.com/jarvisprestidge/domain-web-crawler

# Run the Go installer
RUN go get github.com/onsi/ginkgo/ginkgo \
    go get github.com/onsi/gomega \
    go install github.com/jarvisprestidge/sitemap

WORKDIR /go/bin

# Indicate the binary as our entrypoint
ENTRYPOINT "/bin/ash"
