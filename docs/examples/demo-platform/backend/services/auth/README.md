# Auth Service

Auth microservice for demo-platform platform.

## Development

### Run Locally

```bash
export SERVICE_NAME=auth-service
export PORT=8080
export METRICS_PORT=9090
export LOG_LEVEL=info
export LOG_FORMAT=console

go run cmd/main.go
```

### Build

```bash
go build -o bin/auth cmd/main.go
```

## Configuration

See project root README.md for complete configuration guide.
