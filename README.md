# Test Task Backend

A Go-based backend service for URL analysis.

## Features

- RESTful API endpoints
- Configuration management
- Health check endpoint
- CORS support
- Graceful shutdown
- Structured logging
- Environment-based configuration

## Prerequisites

- Go 1.16 or higher
- Git

## Getting Started

1. Clone the repository:

   ```bash
   git clone https://github.com/ayeshakhan-29/test-task-BE.git
   cd test-task-BE
   ```

2. Copy the example environment file and update it with your configuration:

   ```bash
   cp .env.example .env
   ```

3. Build and run the application:

   ```bash
   # Install dependencies
   go mod download

   # Run the application
   go run cmd/server/main.go
   ```

The server will start on `http://localhost:8080` by default.

## API Endpoints

- `GET /api/v1/health` - Health check endpoint

## Environment Variables

Create a `.env` file in the root directory with the following variables:

```
ENV=development
PORT=8080
APP_VERSION=1.0.0
```

## Project Structure

```
.
├── cmd/                  # Main application entry points
│   └── server/          # Server entry point
├── internal/            # Private application code
│   ├── app/             # Application components
│   │   ├── handlers/    # Request handlers
│   │   ├── models/      # Data models
│   │   ├── repositories/# Data access layer
│   │   └── services/    # Business logic
│   ├── config/          # Configuration
│   └── middleware/      # HTTP middleware
├── pkg/                 # Public libraries
├── api/                 # API contracts
├── configs/             # Configuration files
├── scripts/             # Build and deployment scripts
└── deployments/         # Deployment configurations
```

## Development

### Running Tests

```bash
go test ./...
```

### Building

```bash
go build -o bin/server cmd/server/main.go
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
