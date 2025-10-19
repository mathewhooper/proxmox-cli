.PHONY: help build test lint fmt vet clean install-linters run

# Default target
help:
	@echo "Proxmox CLI - Available targets:"
	@echo "  make build          - Build the proxmox-cli binary"
	@echo "  make test           - Run all tests"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make lint           - Run all linting checks"
	@echo "  make fmt            - Format code with gofmt"
	@echo "  make vet            - Run go vet"
	@echo "  make clean          - Remove build artifacts"
	@echo "  make install-linters - Install all linting tools"
	@echo "  make ci             - Run all CI checks (fmt, lint, vet, test)"
	@echo "  make run            - Build and run the CLI (use ARGS for arguments)"

# Build the binary
build:
	@echo "Building proxmox-cli..."
	go build -v -o proxmox-cli .
	@echo "✓ Build complete: ./proxmox-cli"

# Run all tests
test:
	@echo "Running tests..."
	go test -v ./tests/...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out -covermode=atomic ./...
	@echo "\nCoverage report:"
	go tool cover -func=coverage.out
	@echo "\nTo view HTML coverage report, run: go tool cover -html=coverage.out"

# Run linting
lint:
	@echo "Running linting checks..."
	@chmod +x scripts/lint.sh
	@./scripts/lint.sh

# Format code
fmt:
	@echo "Formatting code..."
	gofmt -w .
	@echo "✓ Code formatted"

# Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...
	@echo "✓ go vet passed"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f proxmox-cli
	rm -f coverage.out coverage.txt
	rm -f test-output.log report.xml
	@echo "✓ Clean complete"

# Install linting tools
install-linters:
	@echo "Installing linting tools..."
	@echo "Installing golangci-lint..."
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.55.2
	@echo "Installing golint..."
	@go install golang.org/x/lint/golint@latest
	@echo "Installing staticcheck..."
	@go install honnef.co/go/tools/cmd/staticcheck@latest
	@echo "Installing errcheck..."
	@go install github.com/kisielk/errcheck@latest
	@echo "Installing gosec..."
	@go install github.com/securego/gosec/v2/cmd/gosec@latest
	@echo "✓ All linters installed"

# Run all CI checks
ci: fmt vet lint test
	@echo ""
	@echo "═══════════════════════════════════════════════════════════"
	@echo "        ✓ ALL CI CHECKS PASSED!"
	@echo "═══════════════════════════════════════════════════════════"

# Build and run
run: build
	./proxmox-cli $(ARGS)

# Quick check before committing
pre-commit: fmt vet lint
	@echo "✓ Pre-commit checks passed. Ready to commit!"
