.PHONY: help build test fmt vet clean run

# Default target
help:
	@echo "Proxmox CLI - Available targets:"
	@echo "  make build          - Build the proxmox-cli binary"
	@echo "  make test           - Run all tests"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make fmt            - Format code with gofmt"
	@echo "  make vet            - Run go vet"
	@echo "  make clean          - Remove build artifacts"
	@echo "  make ci             - Run all CI checks (fmt, vet, test)"
	@echo "  make run            - Build and run the CLI (use ARGS for arguments)"
	@echo ""
	@echo "Linting:"
	@echo "  golangci-lint is used for comprehensive linting"
	@echo "  Run: golangci-lint run"
	@echo "  Install: https://golangci-lint.run/docs/welcome/install/#binaries"

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

# Run all CI checks
ci: fmt vet test
	@echo ""
	@echo "═══════════════════════════════════════════════════════════"
	@echo "        ✓ ALL CI CHECKS PASSED!"
	@echo "═══════════════════════════════════════════════════════════"

# Build and run
run: build
	./proxmox-cli $(ARGS)

# Quick check before committing
pre-commit: fmt vet
	@echo "✓ Pre-commit checks passed. Ready to commit!"
	@echo "Note: Run 'golangci-lint run' for comprehensive linting"
