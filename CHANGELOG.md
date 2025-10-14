# Changelog

All notable changes to EggyByte Core will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.0.1] - 2025-01-27

### Added
- **Initial Release**: First stable release of EggyByte Core v0.0.1
- **Core Bootstrap Functionality**: Single entry point for service initialization
- **Configuration Management**: Environment-based configuration with Kubernetes support
- **Structured Logging**: Context-aware logging with multiple output formats
- **Database Integration**: MySQL/TiDB support with connection pooling and GORM integration
- **Health Check Endpoints**: Kubernetes-compatible health checks (/healthz, /livez, /readyz)
- **Prometheus Metrics**: Metrics exposition (/metrics) with custom collectors
- **Service Lifecycle Management**: Graceful startup and shutdown handling
- **HTTP/gRPC Servers**: Business server implementations with handler registration
- **Comprehensive Documentation**: API reference, architecture guide, and examples
- **Test Coverage**: 67.9% overall test coverage with isolated method testing

### Features
- **Bootstrap System**: Single entry point for service initialization
- **Configuration Management**: Environment-based configuration with Kubernetes support
- **Structured Logging**: Context-aware logging with multiple output formats
- **Database Integration**: MySQL/TiDB support with automatic table migration
- **Service Lifecycle**: Graceful startup and shutdown management
- **Monitoring**: Built-in Prometheus metrics and health checks
- **Documentation**: Comprehensive guides and API reference

### Architecture
- Modular design with clear separation of concerns
- Convention over configuration approach
- Production-ready with built-in observability
- Cloud-native with Kubernetes support
- Developer-friendly with minimal boilerplate

### Supported Platforms
- Go 1.25.1+
- Linux, macOS, Windows
- Docker and Kubernetes
- MySQL 8.0+ and TiDB
- Prometheus monitoring

### Documentation
- [Getting Started Guide](docs/GETTING_STARTED.md)
- [Architecture Guide](docs/ARCHITECTURE.md)
- [API Reference](docs/API_REFERENCE.md)
- [Migration Guide](docs/MIGRATION.md)
- [Contributing Guide](CONTRIBUTING.md)

## [Unreleased]

### Added
- **New Server Module (`pkg/server`)**: HTTP and gRPC business server implementations
  - `HTTPServer` with handler registration and lifecycle management
  - `GRPCServer` with reflection support and custom options
  - `ServerManager` for coordinating multiple servers
  - Comprehensive interfaces for loose coupling and testing
- **Enhanced Monitoring Module (`pkg/monitoring`)**: Split into separate services
  - `HealthService` for health check endpoints (/healthz, /livez, /readyz)
  - `MetricsService` for Prometheus metrics exposition (/metrics)
  - Support for custom health checkers and metrics collectors
- **Enhanced Configuration System**: New environment variables for flexible service control
  - `BUSINESS_HTTP_PORT` (default: 8080) - HTTP API server port
  - `BUSINESS_GRPC_PORT` (default: 9090) - gRPC API server port
  - `HEALTH_CHECK_PORT` (default: 8081) - Health check service port
  - `METRICS_PORT` (default: 9091) - Metrics service port
  - `ENABLE_BUSINESS_HTTP` (default: true) - Enable/disable HTTP server
  - `ENABLE_BUSINESS_GRPC` (default: true) - Enable/disable gRPC server
  - `ENABLE_HEALTH_CHECK` (default: true) - Enable/disable health service
  - `ENABLE_METRICS` (default: true) - Enable/disable metrics service
- **Enhanced Bootstrap Function**: Integrated business server management
  - Conditional service startup based on configuration flags
  - Automatic registration of HTTP/gRPC servers
  - Separate health check and metrics services
  - Improved service lifecycle management
- **Comprehensive Test Coverage**: Added tests for all new modules
  - Server module tests (HTTP, gRPC, interfaces)
  - Monitoring module tests (health, metrics)
  - Bootstrap integration tests
  - Mock implementations for interface testing
- **Comprehensive English Documentation**: Detailed comments and examples
  - All public APIs documented with usage examples
  - Thread safety documentation for concurrent operations
  - Performance notes and best practices
  - Updated API reference with new modules

### Changed
- **BREAKING**: Renamed `Port` field to `BusinessHTTPPort` in Config struct
- **BREAKING**: Split `MonitoringService` into separate `HealthService` and `MetricsService`
- **BREAKING**: Updated Bootstrap function to use new server and monitoring modules
- Enhanced project structure with modern Go library standards
- Improved error handling with nil checks and validation
- Standardized logger interface usage across all modules
- Updated all documentation to reflect new architecture

### Fixed
- Fixed logger interface inconsistencies (`*log.Logger` → `log.Logger`)
- Fixed Prometheus handler imports in metrics tests
- Fixed nil pointer panics in ServerManager
- Fixed compilation errors in test files
- Fixed race conditions in service tests
- Fixed HTTP handler testing issues

### Removed
- Legacy `pkg/monitoring/service.go` (split into health.go and metrics.go)
- Old `pkg/metrics` and `pkg/health` modules (replaced by `pkg/monitoring`)
- Deprecated configuration fields and methods

## [v1.1.0] - 2025-01-01

### Changed
- **BREAKING**: Separated ebcctl CLI tool into standalone repository
- Updated project initialization to create minimal structure with ebc.yaml configuration
- Removed ebcctl-specific build targets from Makefile
- Updated documentation to reference standalone ebcctl repository

### Removed
- ebcctl CLI tool (moved to github.com/eggybyte-technology/ebcctl)
- ebcctl-specific git management commands from Makefile
- ebcctl-specific release management commands from Makefile

## [v1.0.1] - 2025-10-14

### Added
- Enhanced frontend generation with automatic full-stack project detection
- Support for multiple frontend applications in single repository
- Improved frontend directory structure: `frontend/{project-name}/` instead of direct `frontend/`
- Enhanced logging system with color encoding and structured output
- Comprehensive .gitignore updates for generated project files

### Changed
- Frontend generation now creates apps in `frontend/{project-name}/` directory
- Updated project detection to identify full-stack projects and place frontend apps accordingly
- Enhanced project structure documentation and README generation
- Improved Makefile targets for frontend project management

### Fixed
- Frontend directory structure issue in full-stack project generation
- Generated project files now properly excluded from git tracking
- Fixed frontend command to work correctly in full-stack projects

## [v1.0.0] - 2025-10-13

### Added
- GitHub Actions CI/CD workflows for automated testing and releases
- Dependabot configuration for automated dependency updates
- Issue and pull request templates for better community contribution
- Code of Conduct and Security Policy documentation
- Enhanced Makefile with categorized help system and new targets
- Examples directory with basic usage demonstrations
- Testdata directory with sample configurations and logs
- Comprehensive .gitignore for Go projects
- Project structure optimization following GitHub conventions

### Changed
- Optimized directory structure to follow modern Go project conventions
- Enhanced Makefile help system with categorized targets
- Improved project organization with proper file categorization
- Updated documentation structure for better discoverability

### Fixed
- Project structure compliance with GitHub standards
- Makefile target organization and help system
- Documentation structure and accessibility

## [1.0.0] - 2025-01-27

### Added
- Initial release of EggyByte Core v1.0.0
- Core bootstrap functionality with service lifecycle management
- Configuration management with environment variables
- Structured logging with zap integration
- Database integration with GORM and MySQL/TiDB support
- Health check endpoints (/healthz, /livez, /readyz)
- Prometheus metrics exposure (/metrics)
- Service launcher with graceful shutdown
- CLI tool for project scaffolding (now separated to standalone ebcctl repository)
- Flutter frontend project generation with platform selection
- Backend microservice generation with local/GitHub dependency support
- Docker Compose for local development with MySQL
- Comprehensive documentation suite
- GitHub-compliant project structure
- Comprehensive documentation
- Example projects and templates

### Features
- **Bootstrap System**: Single entry point for service initialization
- **Configuration Management**: Environment-based configuration with Kubernetes support
- **Structured Logging**: Context-aware logging with multiple output formats
- **Database Integration**: MySQL/TiDB support with automatic table migration
- **Service Lifecycle**: Graceful startup and shutdown management
- **Monitoring**: Built-in Prometheus metrics and health checks
- **CLI Tool**: Code generation for microservices and repositories
- **Documentation**: Comprehensive guides and API reference

### Architecture
- Modular design with clear separation of concerns
- Convention over configuration approach
- Production-ready with built-in observability
- Cloud-native with Kubernetes support
- Developer-friendly with minimal boilerplate

### Supported Platforms
- Go 1.25.1+
- Linux, macOS, Windows
- Docker and Kubernetes
- MySQL 8.0+ and TiDB
- Memcached 1.6+

### Documentation
- [Getting Started Guide](docs/GETTING_STARTED.md)
- [Architecture Guide](docs/ARCHITECTURE.md)
- [API Reference](docs/API_REFERENCE.md)
- [Migration Guide](docs/MIGRATION.md)
- [Contributing Guide](CONTRIBUTING.md)

### Examples
- [Demo Platform](docs/examples/demo-platform/)
- [Microservice Examples](docs/examples/EXAMPLES.md)
- [Configuration Templates](configs/templates/)
- [Docker Deployment](deployments/docker/)
- [Kubernetes Deployment](deployments/kubernetes/)

## [0.9.0] - 2025-01-20

### Added
- Initial development version
- Basic core functionality
- Database integration
- Logging system
- Configuration management

### Changed
- Multiple iterations and improvements
- API refinements
- Performance optimizations

### Fixed
- Various bugs and issues
- Configuration loading problems
- Database connection issues

## [0.8.0] - 2025-01-15

### Added
- Service lifecycle management
- Health check system
- Monitoring integration
- CLI tool foundation

### Changed
- Improved error handling
- Enhanced logging
- Better configuration support

### Fixed
- Service startup issues
- Health check problems
- Configuration validation

## [0.7.0] - 2025-01-10

### Added
- Database abstraction layer
- Repository pattern
- Auto-registration system
- Connection pooling

### Changed
- Improved database integration
- Better error handling
- Enhanced configuration

### Fixed
- Database connection issues
- Migration problems
- Configuration loading

## [0.6.0] - 2025-01-05

### Added
- Configuration management
- Environment variable support
- Kubernetes ConfigMap watching
- Thread-safe configuration access

### Changed
- Improved configuration loading
- Better error handling
- Enhanced validation

### Fixed
- Configuration race conditions
- Environment variable parsing
- Validation issues

## [0.5.0] - 2024-12-30

### Added
- Structured logging system
- Context propagation
- Multiple output formats
- Request ID tracking

### Changed
- Improved logging performance
- Better context handling
- Enhanced error reporting

### Fixed
- Logging performance issues
- Context propagation problems
- Error reporting bugs

## [0.4.0] - 2024-12-25

### Added
- Core bootstrap system
- Service lifecycle management
- Graceful shutdown handling
- Signal handling

### Changed
- Improved service management
- Better error handling
- Enhanced startup process

### Fixed
- Service startup issues
- Shutdown problems
- Signal handling bugs

## [0.3.0] - 2024-12-20

### Added
- Basic service framework
- HTTP server support
- gRPC server support
- Service interfaces

### Changed
- Improved service architecture
- Better error handling
- Enhanced API design

### Fixed
- Service interface issues
- HTTP server problems
- gRPC server bugs

## [0.2.0] - 2024-12-15

### Added
- Project structure
- Basic modules
- Initial documentation
- Build system

### Changed
- Improved project organization
- Better module structure
- Enhanced documentation

### Fixed
- Build issues
- Module dependencies
- Documentation problems

## [0.1.0] - 2024-12-10

### Added
- Initial project setup
- Basic structure
- Initial documentation
- Version control

### Changed
- Project initialization
- Basic setup
- Initial configuration

### Fixed
- Initial setup issues
- Configuration problems
- Documentation errors

---

## Release Notes

### Version 1.0.0

This is the first stable release of EggyByte Core, providing a comprehensive foundation for building Go microservices.

#### Key Features

1. **Enterprise-Grade Foundation**
   - Production-ready microservice framework
   - Built-in observability and monitoring
   - Comprehensive error handling and logging

2. **Developer Experience**
   - Minimal boilerplate with single bootstrap call
   - Powerful CLI tool for code generation
   - Comprehensive documentation and examples

3. **Cloud-Native Design**
   - Kubernetes-ready with health checks
   - Environment-based configuration
   - Graceful shutdown handling

4. **Modular Architecture**
   - Clear separation of concerns
   - Easy to extend and customize
   - Convention over configuration

#### Getting Started

```bash
# Install CLI tool (now standalone)
go install github.com/eggybyte-technology/ebcctl@latest

# Create new service
ebcctl init backend my-service
cd my-service

# Run service
go run cmd/main.go
```

#### Documentation

- [Getting Started Guide](docs/GETTING_STARTED.md)
- [Architecture Guide](docs/ARCHITECTURE.md)
- [API Reference](docs/API_REFERENCE.md)
- [Migration Guide](docs/MIGRATION.md)

#### Examples

- [Demo Platform](docs/examples/demo-platform/)
- [Microservice Examples](docs/examples/EXAMPLES.md)

#### Support

- **GitHub Issues**: [Report bugs](https://github.com/eggybyte-technology/go-eggybyte-core/issues)
- **GitHub Discussions**: [Ask questions](https://github.com/eggybyte-technology/go-eggybyte-core/discussions)
- **Email**: support@eggybyte.com

#### License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

#### Acknowledgments

- Thanks to all contributors who helped make this release possible
- Special thanks to the Go community for excellent tools and libraries
- Appreciation to the open source community for inspiration and best practices

---

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

## Support

- **Documentation**: [docs/](docs/)
- **Issues**: [GitHub Issues](https://github.com/eggybyte-technology/go-eggybyte-core/issues)
- **Discussions**: [GitHub Discussions](https://github.com/eggybyte-technology/go-eggybyte-core/discussions)
- **Email**: support@eggybyte.com

## License

Copyright © 2025 EggyByte Technology. All rights reserved.

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.