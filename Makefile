.PHONY: build test clean install lint fmt

# Build the CLI
build:
	go build -o bin/difftron ./cmd/difftron

# Run all tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Install the CLI
install:
	go install ./cmd/difftron

# Format code
fmt:
	go fmt ./...

# Run linter (requires golangci-lint)
lint:
	golangci-lint run

# Clean build artifacts
clean:
	rm -rf bin/ coverage.out coverage.html

# Run the CLI locally (example)
run:
	go run ./cmd/difftron
