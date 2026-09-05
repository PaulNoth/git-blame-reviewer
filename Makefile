.PHONY: build test lint clean help install-tools

# Build the binary
build:
	go build -o git-blame-reviewer .

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run linter
lint:
	golangci-lint run

# Clean build artifacts
clean:
	rm -f git-blame-reviewer
	rm -f coverage.out
	rm -f coverage.html

# Install development tools
# Pinned (not @latest): golangci-lint v1.x can't read the export data
# produced by newer Go toolchains, and .golangci.yml uses the v2 config
# schema, so an unpinned install can silently resolve to an incompatible
# major version. Bump this deliberately, together with the `version` input
# in .github/workflows/ci.yml, when upgrading.
install-tools:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.13.2

# Run all checks (test + lint)
check: test lint

# Display help
help:
	@echo "Available targets:"
	@echo "  build         - Build the binary"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  lint          - Run linter"
	@echo "  clean         - Clean build artifacts"
	@echo "  install-tools - Install development tools"
	@echo "  check         - Run tests and linter"
	@echo "  help          - Show this help"