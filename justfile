# azuredevops-go development tasks

# Run all tests
test:
    go test ./... -v

# Run tests with race detection
test-race:
    go test ./... -v -race

# Run linter
lint:
    golangci-lint run ./...

# Run go vet
vet:
    go vet ./...

# Format code
fmt:
    gofmt -w .
    goimports -w .

# Tidy dependencies
tidy:
    go mod tidy

# Run all checks (test + lint + vet)
check: fmt vet lint test

# Show test coverage
cover:
    go test ./... -coverprofile=coverage.out
    go tool cover -html=coverage.out -o coverage.html
    @echo "Coverage report: coverage.html"

# Show package documentation
doc:
    go doc ./webhooks/...
