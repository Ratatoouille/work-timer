.PHONY: build test run lint clean fmt vet

BINARY = work-timer
VERSION = $(shell grep 'const version' main.go | awk -F'"' '{print $$2}')

build:
	go build -o $(BINARY) .

test:
	go test ./...

run:
	go run .

fmt:
	go fmt ./...

vet:
	go vet ./...

lint: vet
	@command -v golangci-lint >/dev/null 2>&1 && golangci-lint run ./... || echo "golangci-lint not installed, skipping"

clean:
	rm -f $(BINARY) dist/

release:
	@if command -v goreleaser >/dev/null 2>&1; then \
		goreleaser release --clean --snapshot; \
	else \
		echo "goreleaser not installed"; exit 1; \
	fi

release-snapshot:
	@if command -v goreleaser >/dev/null 2>&1; then \
		goreleaser release --clean --snapshot; \
	else \
		echo "goreleaser not installed"; exit 1; \
	fi
