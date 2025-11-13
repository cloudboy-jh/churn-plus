.PHONY: build run clean test install

# Binary name
BINARY=churn-plus
VERSION=0.1.0

# Build the binary
build:
	go build -ldflags "-X main.version=$(VERSION)" -o $(BINARY) ./cmd/churn-plus

# Run locally
run:
	go run ./cmd/churn-plus

# Clean build artifacts
clean:
	rm -f $(BINARY)
	rm -rf dist/

# Run tests
test:
	go test -v ./...

# Install to $GOPATH/bin
install:
	go install -ldflags "-X main.version=$(VERSION)" ./cmd/churn-plus

# Build for multiple platforms
build-all:
	mkdir -p dist
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o dist/$(BINARY)-linux-amd64 ./cmd/churn-plus
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o dist/$(BINARY)-darwin-amd64 ./cmd/churn-plus
	GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.version=$(VERSION)" -o dist/$(BINARY)-darwin-arm64 ./cmd/churn-plus
	GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o dist/$(BINARY)-windows-amd64.exe ./cmd/churn-plus

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Show help
help:
	@echo "Churn-Plus Makefile"
	@echo ""
	@echo "Targets:"
	@echo "  build      - Build the binary"
	@echo "  run        - Run locally"
	@echo "  clean      - Clean build artifacts"
	@echo "  test       - Run tests"
	@echo "  install    - Install to \$$GOPATH/bin"
	@echo "  build-all  - Build for multiple platforms"
	@echo "  fmt        - Format code"
	@echo "  lint       - Lint code"
	@echo "  help       - Show this help"
