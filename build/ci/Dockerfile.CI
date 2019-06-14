FROM golang:1.12

ENV GOLANGCI_LINT_VERSION v1.15.0

# Dependencies
RUN wget -O - -q https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s ${GOLANGCI_LINT_VERSION}