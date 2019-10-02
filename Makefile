all: container

VERSION = 0.0.1
TAG = $(VERSION)
PREFIX = nginx/crossplane-go

DOCKER_RUN = docker run --rm -v $(shell pwd):/go/src/gitswarm.f5net.com/indigo/poc/crossplane-go
DOCKER_BUILD_RUN = docker run --rm -v $(shell pwd):/go/src/gitswarm.f5net.com/indigo/poc/crossplane-go -w /go/src/gitswarm.f5net.com/indigo/poc/crossplane-go
BUILD_IN_CONTAINER = 1
DOCKERFILEPATH = build
GOLANG_CONTAINER = golang:1.12

requirements:
	go get -u \
    github.com/golang/dep/cmd/dep \
    github.com/golangci/golangci-lint/cmd/golangci-lint

dependencies:
	dep ensure

build:
ifeq ($(BUILD_IN_CONTAINER),1)
	$(DOCKER_BUILD_RUN) -e CGO_ENABLED=0 $(GOLANG_CONTAINER) go build -installsuffix cgo -ldflags "-w" -o /go/src/gitswarm.f5net.com/indigo/poc/crossplane-go/crossplane-go
else
	CGO_ENABLED=0 GOOS=linux go build -installsuffix cgo -ldflags "-w" -o crossplane-go gitswarm.f5net.com/indigo/poc/crossplane-go/cmd/crossplane.go
endif

test:
ifeq ($(BUILD_IN_CONTAINER),1)
	docker run --rm -v $(shell pwd):/go/src/gitswarm.f5net.com/indigo/poc/crossplane-go \
	$(shell docker build -f ./build/Dockerfile -q .) \
	go test $(shell go list ./... | grep -v /vendor/)
else
	go test ./...
endif

lint:
	golangci-lint run

clean:
	rm -f crossplane-go
