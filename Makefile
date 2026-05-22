.PHONY: deps build test dev

APP_NAME := tld
BUILD_DIR := bin
ROOT_DIR := $(shell pwd)

deps:
	go mod download

build:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/$(APP_NAME)

test:
	go test ./...

dev:
	osascript \
		-e 'tell application "iTerm"' \
		-e 'activate' \
		-e 'if (count of windows) = 0 then' \
		-e 'set devWindow to (create window with default profile)' \
		-e 'else' \
		-e 'set devWindow to current window' \
		-e 'tell devWindow to create tab with default profile' \
		-e 'end if' \
		-e 'tell current session of devWindow' \
		-e 'write text "cd $(ROOT_DIR) && go run ./cmd/$(APP_NAME)"' \
		-e 'end tell' \
		-e 'end tell'
