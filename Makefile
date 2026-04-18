.PHONY: test test-race test-cover test-verbose test-faults test-metrics vet build run clean

## Run all tests
test:
	go test ./...

## Run all tests with race detector (catches concurrent access bugs)
test-race:
	go test -race ./...

## Run tests with coverage report — opens HTML in browser
test-cover:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## Run tests with verbose output (shows each test name and result)
test-verbose:
	go test -v -race ./...

## Run only fault injection tests
test-faults:
	go test -v -race ./internal/domain/faults/...

## Run only metrics tests
test-metrics:
	go test -v -race ./internal/metrics/...

## Run only router integration tests
test-router:
	go test -v -race ./internal/router/...

## Run only e2e multi-fault scenario tests
test-e2e:
	go test -v -race -run TestE2E ./internal/router/...

## Static analysis
vet:
	go vet ./...

## Build the binary
build:
	go build -o bin/glitchmesh ./cmd/glitchmesh

## Run the server
run: build
	./bin/glitchmesh start server

## Clean build artifacts
clean:
	rm -rf bin/ coverage.out coverage.html
