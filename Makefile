# note: call scripts from /scripts
GO_MAIN=./cmd/tld/tld.go
GO_TEST=./test

install:
	go install $(GO_MAIN)

run:
	go run $(GO_MAIN)

build:
	go build $(GO_MAIN)

test:
	go test ${GO_TEST}

format:
	go fmt $(GO_MAIN)
	go fmt $(GO_TEST)
