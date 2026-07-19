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
	go test -v -race ./internal/dataplane/faults/...

## Run only metrics tests
test-metrics:
	go test -v -race ./internal/dataplane/metrics/...

## Run only router integration tests
test-router:
	go test -v -race ./internal/dataplane/server/...

## Run only e2e multi-fault scenario tests
test-e2e:
	go test -v -race -run TestE2E ./internal/dataplane/server/...

## Static analysis
vet:
	go vet ./...

## Build binaries
build:
	go build -o bin/glitchmesh ./cmd/glitchmesh

## Run the server
run: build
	./bin/glitchmesh start server

## Clean build artifacts
clean:
	rm -rf bin/ coverage.out coverage.html

## Local distributed lab (Docker Compose)
lab-up:
	cd lab && docker compose up --build

lab-down:
	cd lab && docker compose down

lab-traffic:
	go run ./cmd/trafficgen/main.go -url http://localhost:8080/api/feed -concurrency 5 -count 50

## Phase 1: Admin API tests only
test-adminapi:
	go test -v -race ./internal/controlplane/adminapi/...

## Phase 1: Config hot-reload + override layer tests
test-config:
	go test -v -race ./internal/controlplane/config/...

## Phase 1: Validation model tests
test-models:
	go test -v -race ./internal/shared/models/...

## Phase 1: Hot-reload e2e tests
test-hotreload:
	go test -v -race -run TestHotReload ./internal/dataplane/server/...

## Phase 1: All Phase 1 tests together
test-phase1:
	go test -v -race ./internal/shared/models/... ./internal/controlplane/config/... ./internal/controlplane/adminapi/... ./internal/dataplane/server/...

## Phase 1: curl example — list services (set ADMIN_TOKEN env var first if auth is enabled)
admin-demo:
	@echo "=== GET /admin/services ==="
	curl -s -H "Authorization: Bearer $${ADMIN_TOKEN}" http://localhost:9000/admin/services | python3 -m json.tool || true
	@echo ""
	@echo "=== GET /admin/config/diff ==="
	curl -s -H "Authorization: Bearer $${ADMIN_TOKEN}" http://localhost:9000/admin/config/diff | python3 -m json.tool || true

