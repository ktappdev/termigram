BINARY := termigram
VERSION ?= $(shell git describe --tags --dirty 2>/dev/null || echo "dev")

.PHONY: build build-version test lint coverage clean

build:
	TELEGRAM_APP_ID_BAKED=$(TELEGRAM_APP_ID_BAKED) TELEGRAM_APP_HASH_BAKED=$(TELEGRAM_APP_HASH_BAKED) ./build.sh

build-version:
	VERSION=$(VERSION) BINARY_NAME=$(BINARY) TELEGRAM_APP_ID_BAKED=$(TELEGRAM_APP_ID_BAKED) TELEGRAM_APP_HASH_BAKED=$(TELEGRAM_APP_HASH_BAKED) ./build.sh

test:
	go test -race -coverprofile=coverage.out -covermode=atomic ./...

lint:
	golangci-lint run

coverage: test
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

clean:
	rm -f coverage.out coverage.html
	rm -f $(BINARY)
