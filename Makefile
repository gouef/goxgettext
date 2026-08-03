.PHONY: install test tests coverage coverage-html build release

GO ?= /usr/local/go/bin/go
GOROOT_PATH ?= /usr/local/go
BINARY_NAME ?= goxgettext
BUILD_DIR ?= bin
DIST_DIR ?= dist

install:
	GOROOT=$(GOROOT_PATH) $(GO) mod tidy

build:
	mkdir -p $(BUILD_DIR)
	GOROOT=$(GOROOT_PATH) $(GO) build -o $(BUILD_DIR)/$(BINARY_NAME) .

release:
	mkdir -p $(DIST_DIR)
	GOROOT=$(GOROOT_PATH) GOOS=linux GOARCH=amd64 $(GO) build -trimpath -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 .
	GOROOT=$(GOROOT_PATH) GOOS=darwin GOARCH=amd64 $(GO) build -trimpath -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 .
	GOROOT=$(GOROOT_PATH) GOOS=windows GOARCH=amd64 $(GO) build -trimpath -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe .

test tests:
	GOROOT=$(GOROOT_PATH) $(GO) test ./...

coverage:
	GOROOT=$(GOROOT_PATH) $(GO) test -covermode=set -coverpkg=./... -coverprofile=coverage.txt ./... && GOROOT=$(GOROOT_PATH) $(GO) tool cover -func=coverage.txt

coverage-html:
	GOROOT=$(GOROOT_PATH) $(GO) test -covermode=set -coverpkg=./... -coverprofile=coverage.txt ./... && GOROOT=$(GOROOT_PATH) $(GO) tool cover -html=coverage.txt -o coverage.html