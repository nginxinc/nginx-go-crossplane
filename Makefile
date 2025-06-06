PACKAGE           = $(notdir $(patsubst %/,%,$(dir $(realpath $(lastword $(MAKEFILE_LIST))))))
OUT_DIR          ?= build
RESULTS_DIR		 ?= results
DOCKER_REGISTRY  ?= local
DOCKER_TAG       ?= latest
LINT_BIN := ./bin/golangci-lint

SHELL=/bin/bash
.SHELLFLAGS=-c -eo pipefail

#######################################
## Local set up.
#######################################

.PHONY: init deps deps-upgrade fmt test lint lint-shell gen

init:
	git config core.hooksPath .githooks
	go install golang.org/x/tools/cmd/goimports
	go install github.com/jstemmer/go-junit-report

deps:
	go mod download
	go mod tidy
	go mod verify

deps-upgrade:
	GOFLAGS="" go get -u ./...
	$(MAKE) deps

#######################################
## Tests, codegen, lint and format.
#######################################
fmt:
	$(info Running goimports...)
	@goimports -w -e $$(find . -type f -name '*.go')

test:
	$(info Running unit tests...)
	mkdir -p $(RESULTS_DIR)
	CGO_ENABLED=1 go test -race -v -cover ./... -coverprofile=$(RESULTS_DIR)/$(PACKAGE)-coverage.out 2>&1 | tee >(go-junit-report > $(RESULTS_DIR)/report.xml)
	@echo "Total code coverage:"
	@go tool cover -func=$(RESULTS_DIR)/$(PACKAGE)-coverage.out | grep 'total:' | tee $(RESULTS_DIR)/anybadge.out
	@go tool cover -html=$(RESULTS_DIR)/$(PACKAGE)-coverage.out -o $(RESULTS_DIR)/coverage.html

test-only-failed:
	$(info Running unit tests (showing only failed ones with context)...)
	go test -v -race ./... | grep --color -B 45 -A 5 -E '^FAIL.+'

$(LINT_BIN):
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.51.2

lint: $(LINT_BIN)
	$(LINT_BIN) run

lint-docker:
	docker run --rm -v "${PWD}":/app -w /app golangci/golangci-lint:v1.51.2 golangci-lint run

lint-shell:
	shellcheck -x $$(find . -name "*.sh" -type f)

gen:
	go generate -x ./...
	$(MAKE) fmt
