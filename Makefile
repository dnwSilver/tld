.PHONY: deps build install test dev

APP_NAME := tld
BUILD_DIR := bin
ROOT_DIR := $(shell pwd)
INSTALL_DIR ?= $(HOME)/.local/bin

deps:
	go mod download

build:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/$(APP_NAME)

install: build
	mkdir -p $(INSTALL_DIR)
	install -m 0755 $(BUILD_DIR)/$(APP_NAME) $(INSTALL_DIR)/$(APP_NAME)
	@echo "Installed $(APP_NAME) to $(INSTALL_DIR)/$(APP_NAME)"
	@echo "Make sure $(INSTALL_DIR) is in your PATH"

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
