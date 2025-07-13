.PHONY: build run test clean migrate-up migrate-down docker-build docker-run docker-down

# Build the application
build:
	go build -o bin/server ./cmd/server

# Run the application
run:
	go run cmd/server/main.go

# Run tests
test:
	go test -v -cover -coverprofile=coverage.out ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Run database migrations
migrate-up:
	go run cmd/migrate/main.go up

# Rollback the last migration
migrate-down:
	go run cmd/migrate/main.go down

# Build Docker image
docker-build:
	docker-compose build

# Start the application with Docker Compose
docker-up:
	docker-compose up -d

# Stop and remove containers
docker-down:
	docker-compose down

# Show logs
docker-logs:
	docker-compose logs -f

# Run tests in Docker
docker-test:
	docker-compose run --rm app go test -v ./...

# Access the app container
docker-bash:
	docker-compose exec app sh

# Access the database
db-connect:
	docker-compose exec db mysql -u root -psecret test_task

# Format code
format:
	go fmt ./...

# Run linter
lint:
	golangci-lint run

# Run linter with fix
lint-fix:
	golangci-lint run --fix

# Generate API documentation
docs:
	swag init -g cmd/server/main.go
