# note: call scripts from /scripts
GO_MAIN=./cmd/tld/tld.go
GO_TEST=./test

## install: Install missing dependencies.
install:
	go install $(GO_MAIN)

## Run application.
run:
	go run $(GO_MAIN)

## Build application.
build:
	go build $(GO_MAIN)

## Run units tests.
test:
	go test $(GO_TEST)

## Format code style.
format:
	go fmt $(GO_MAIN)
	go fmt $(GO_TEST)

.PHONY: install run build test format
