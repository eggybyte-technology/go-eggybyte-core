# Basic Example

This example demonstrates the minimal setup required to create a microservice using EggyByte Core.

## What This Example Shows

- Basic service configuration
- Service lifecycle management
- Graceful shutdown handling
- Structured logging

## Running the Example

1. Set environment variables:
   ```bash
   export SERVICE_NAME=basic-example
   export PORT=8080
   export LOG_LEVEL=info
   export LOG_FORMAT=console
   ```

2. Run the service:
   ```bash
   go run main.go
   ```

3. Check health endpoints:
   ```bash
   curl http://localhost:9090/healthz
   curl http://localhost:9090/metrics
   ```

## Key Features Demonstrated

- **Zero Configuration**: Service starts with minimal setup
- **Graceful Shutdown**: Handles SIGTERM and SIGINT signals
- **Health Checks**: Built-in health endpoints
- **Structured Logging**: Context-aware logging with request IDs
- **Monitoring**: Prometheus metrics exposed automatically

## Next Steps

- Try the [Database Example](../database/) to see database integration
- Check the [Advanced Example](../advanced/) for custom health checkers
- Explore the [CLI Tool](../../cmd/ebcctl/) for code generation
