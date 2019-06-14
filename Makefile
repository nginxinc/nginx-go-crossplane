all: container

VERSION = 0.0.1
TAG = $(VERSION)
PREFIX = nginx/crossplane-go

DOCKER_RUN = docker run --rm -v $(shell pwd):/go/src/github.com/nginxinc/crossplane-go
DOCKER_BUILD_RUN = docker run --rm -v $(shell pwd):/go/src/github.com/nginxinc/crossplane-go -w /go/src/github.com/nginxinc/crossplane-go
BUILD_IN_CONTAINER = 1
DOCKERFILEPATH = build
GOLANG_CONTAINER = golang:1.12

build:
ifeq ($(BUILD_IN_CONTAINER),1)
	$(DOCKER_BUILD_RUN) -e CGO_ENABLED=0 $(GOLANG_CONTAINER) go build -installsuffix cgo -ldflags "-w" -o /go/src/github.com/nginxinc/crossplane-go/crossplane-go
else
	CGO_ENABLED=0 GOOS=linux go build -installsuffix cgo -ldflags "-w" -o crossplane-go github.com/nginxinc/crossplane-go/cmd/crossplane.go
endif

test:
ifeq ($(BUILD_IN_CONTAINER),1)
	$(DOCKER_RUN) $(GOLANG_CONTAINER) go test ./...
else
	go test ./...
endif

lint:
	golangci-lint run

clean:
	rm -f crossplane-go
	rm -f Dockerfile