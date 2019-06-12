init:
ifeq ($(UNAME_S),Linux)
	curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
endif
ifeq ($(UNAME_S),Darwin)
	brew upgrade dep
	brew install dep
endif
build:
	go build -o crossplane
test:
	go test ./..
ftest:
	echo "functional tests not implemented yet"
	exit 1
