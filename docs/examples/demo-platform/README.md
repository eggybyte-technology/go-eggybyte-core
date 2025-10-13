# Demo Platform

Full-stack application built with EggyByte technology stack.

## Project Structure

- `backend/` - Backend microservices (Go)
  - `services/auth/` - Authentication service
  - `services/user/` - User management service
- `frontend/` - Flutter application (Android/iOS/Web)
- `api/` - Shared API definitions (Protocol Buffers)
- `scripts/` - Build and deployment scripts

## Prerequisites

### Backend
- Go 1.25.1+
- Docker and Docker Compose
- Make

### Frontend
- Flutter SDK 3.16.0+
- Dart 3.2.0+

## Quick Start

### 1. Start Backend Services

```bash
# Start all backend services with Docker Compose
make dev-up

# Or run individual services locally
cd backend/services/auth
go run cmd/main.go
```

### 2. Run Frontend

```bash
cd frontend
flutter pub get
flutter run
```

## Development

### Backend Development

Each service is an independent Go module:

```bash
cd backend/services/user
go run cmd/main.go
```

### Frontend Development

```bash
cd frontend
flutter run -d chrome  # Run on web
flutter run            # Run on mobile device/emulator
```

### Generate Repository Code

```bash
cd backend/services/user
ebcctl new repo <model-name>
```

## Build & Deploy

### Build All Services

```bash
make build-all
```

### Build Docker Images

```bash
make docker-build-all
```

### Deploy to Kubernetes

```bash
make deploy-dev
```

## Configuration

### Backend Services

Each service is configured via environment variables. See individual service README.md files.

Common variables:
- `SERVICE_NAME` - Service identifier
- `PORT` - HTTP server port
- `DATABASE_DSN` - Database connection string
- `LOG_LEVEL` - Logging level (debug, info, warn, error)

### Frontend

Configure API endpoints in `frontend/lib/config/api_config.dart`.

## Testing

### Backend Tests

```bash
make test-backend
```

### Frontend Tests

```bash
cd frontend
flutter test
```

## License

Copyright © 2025 EggyByte Technology
