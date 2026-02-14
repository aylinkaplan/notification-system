.PHONY: run build test lint docker-up docker-down migrate test-api

run:
	go run ./cmd/api

build:
	go build -o bin/api ./cmd/api

test:
	go test ./...

lint:
	golangci-lint run

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

migrate:
	@echo "Migrations run automatically on startup"

test-api:
	@bash scripts/test-api.sh
