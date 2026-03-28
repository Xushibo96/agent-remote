SHELL := /bin/bash

PROJECT := agent-remote
DIST_DIR ?= dist
GO ?= go

.PHONY: build build-linux build-darwin package release test clean

build:
	./scripts/build.sh

build-linux:
	GOOS=linux GOARCH=amd64 ./scripts/build.sh

build-darwin:
	GOOS=darwin GOARCH=arm64 ./scripts/build.sh

package:
	./scripts/package.sh

release:
	./scripts/release.sh

test:
	$(GO) test ./...

clean:
	rm -rf "$(DIST_DIR)"
