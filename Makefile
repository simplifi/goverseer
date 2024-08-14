TAG?=""

.DEFAULT_GOAL := test

# Run all tests
.PHONY: test
test: fmt vet test-unit tidy

# Run unit tests
.PHONY: test-unit
test-unit:
	go test -count 1 -v -race ./...

# Clean go.mod
.PHONY: tidy
tidy:
	go mod tidy
	git diff --exit-code go.sum

# Check formatting
.PHONY: fmt
fmt:
	test -z "$(shell gofmt -l .)"

# Run vet
.PHONY: vet
vet:
	go vet ./...

# Clean up any cruft left over from old builds
.PHONY: clean
clean:
	rm -rf goverseer dist/

# Build the application
.PHONY: build
build: clean
	CGO_ENABLED=0 go build ./cmd/goverseer

# Create a git tag
.PHONY: tag
tag:
	git tag -a $(TAG) -m "Release $(TAG)"
	git push origin $(TAG)
