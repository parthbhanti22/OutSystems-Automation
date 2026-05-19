BINARY := os-agent
VERSION := 0.1.0
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: build build-linux-amd64 build-linux-arm64 clean test

## build: Compile a statically linked binary for the current platform.
build:
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o $(BINARY) .

## build-linux-amd64: Cross-compile for Linux x86_64.
build-linux-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BINARY)-linux-amd64 .

## build-linux-arm64: Cross-compile for Linux ARM64 (e.g., Raspberry Pi).
build-linux-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(BINARY)-linux-arm64 .

## clean: Remove compiled binaries and test artifacts.
clean:
	rm -f $(BINARY) $(BINARY)-*
	rm -rf test-workspace/

## test: Run all unit tests.
test:
	go test ./... -v -race
