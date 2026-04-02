.PHONY: build run test lint clean fmt tidy vet scrape

BINARY   = scraper
BUILD_DIR = bin

## build: Compile the scraper binary
build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY) ./cmd/scraper

## run: Run the scraper (pass ARGS, e.g. make run ARGS="https://example.com --engine colly")
run:
	go run ./cmd/scraper $(ARGS)

## test: Run all tests with race detection
test:
	go test ./... -v -race -count=1

## lint: Run golangci-lint
lint:
	golangci-lint run ./...

## fmt: Format all Go source files
fmt:
	gofmt -w .

## tidy: Tidy go.mod and go.sum
tidy:
	go mod tidy

## vet: Run go vet
vet:
	go vet ./...

## clean: Remove build artifacts
clean:
	rm -rf $(BUILD_DIR)

## scrape: Build and scrape a URL (pass URL, e.g. make scrape URL=https://example.com)
scrape: build
	./$(BUILD_DIR)/$(BINARY) $(URL)

## help: Show this help
help:
	@echo "Available targets:"
	@grep -E '^## ' Makefile | sed 's/## /  /'
