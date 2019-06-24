FROM golang:1

RUN go get -u \
    github.com/golang/dep/cmd/dep \
    github.com/golangci/golangci-lint/cmd/golangci-lint

# for head-to-head comparison tests (see main_test)
RUN apt-get -y update && \
    apt-get -y install python3 python3-pip && \
    pip3 install crossplane


WORKDIR /go/src/github.com/nginxinc/crossplane-go