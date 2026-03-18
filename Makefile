BINARY := termigram

.PHONY: build build-version

build:
	./build.sh

build-version:
	VERSION=$(VERSION) BINARY_NAME=$(BINARY) ./build.sh
