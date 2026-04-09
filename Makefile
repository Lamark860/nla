.PHONY: build run test test-v test-cover lint clean docker-up docker-down docker-rebuild docker-logs

# Build
build:
	go build -o bin/nla-api ./cmd/api

run: build
	./bin/nla-api

# Tests
test:
	go test ./internal/... -count=1

test-v:
	go test ./internal/... -v -count=1

test-cover:
	go test ./internal/... -coverprofile=coverage.out -count=1
	go tool cover -func=coverage.out
	@echo ""
	@echo "HTML report: go tool cover -html=coverage.out"

# Docker
docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-rebuild:
	docker compose up -d --build api

docker-logs:
	docker compose logs -f api

docker-test:
	docker compose exec api go test ./internal/... -v -count=1

# Cleanup
clean:
	rm -rf bin/ coverage.out

# Dev helpers
deps:
	go mod tidy

lint:
	@which golangci-lint > /dev/null 2>&1 || echo "Install: brew install golangci-lint"
	golangci-lint run ./...
