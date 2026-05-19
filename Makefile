.PHONY: up down migrate seed test build run lint clean

# Start everything
up:
	docker compose up --build -d
	@echo "API running at http://localhost:8080"
	@echo "Swagger docs at http://localhost:8080/swagger/"

# Stop everything
down:
	docker compose down

# View logs
logs:
	docker compose logs -f

# Run tests
test:
	go test ./... -v -count=1 -timeout=300s

# Build binaries
build:
	go build -ldflags="-s -w" -o bin/api ./cmd/api
	go build -ldflags="-s -w" -o bin/seed ./cmd/seed

# Lint
lint:
	golangci-lint run --timeout=5m

# Generate swagger docs
swagger:
	swag init -g cmd/api/main.go --output docs

# Rebuild API without cache
rebuild:
	docker compose build --no-cache api
	docker compose up -d api

# Run migration + seed manually (for local dev without docker)
migrate:
	@echo "Run: migrate -path migrations -database \"$(DATABASE_URL)\" up"

seed:
	go run ./cmd/seed

run:
	go run ./cmd/api

# Clean build artifacts
clean:
	rm -rf bin/
	docker compose down -v
