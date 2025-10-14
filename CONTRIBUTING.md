# Contributing to EggyByte Core

Thank you for your interest in contributing to EggyByte Core! This document provides guidelines and information for contributors.

## 📋 Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Contributing Process](#contributing-process)
- [Coding Standards](#coding-standards)
- [Testing Guidelines](#testing-guidelines)
- [Documentation](#documentation)
- [Release Process](#release-process)

## 🤝 Code of Conduct

This project follows the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md). By participating, you agree to uphold this code.

## 🚀 Getting Started

### Prerequisites

- Go 1.25.1 or later
- Git
- Make (for build automation)
- Docker (optional, for testing)

### Development Setup

1. **Fork and Clone**
   ```bash
   git clone https://github.com/your-username/go-eggybyte-core.git
   cd go-eggybyte-core
   ```

2. **Install Dependencies**
   ```bash
   go mod download
   ```

3. **Run Tests**
   ```bash
   make test
   ```

4. **Run Linting**
   ```bash
   make lint
   ```

5. **Full Check**
   ```bash
   make check
   ```

## 🔄 Contributing Process

### 1. Create an Issue

Before starting work, please:
- Check existing issues to avoid duplicates
- Create an issue describing the problem or feature
- Wait for maintainer approval before starting work

### 2. Create a Branch

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/your-bug-fix
```

### 3. Make Changes

- Follow our [coding standards](#coding-standards)
- Write tests for new functionality
- Update documentation as needed
- Ensure all tests pass

### 4. Commit Changes

Use conventional commit messages:

```bash
git commit -m "feat: add new health check endpoint"
git commit -m "fix: resolve database connection timeout"
git commit -m "docs: update API documentation"
```

### 5. Push and Create PR

```bash
git push origin feature/your-feature-name
```

Then create a Pull Request with:
- Clear description of changes
- Reference to related issues
- Screenshots (if applicable)
- Test results

## 📝 Coding Standards

### Go Code Style

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` and `goimports` for formatting
- Maximum 50 lines per public method
- Maximum 100 lines per file (exceptions allowed with justification)
- Use meaningful variable and function names
- Write comprehensive comments for public APIs

### Package Organization

```
pkg/
├── core/           # Core bootstrap and initialization
├── config/         # Configuration management
├── log/            # Structured logging
├── db/             # Database integration
├── server/         # HTTP/gRPC servers
├── monitoring/     # Health checks and metrics
└── service/        # Service lifecycle management
```

### Error Handling

- Use `error` interface, not `commonpb.ErrorInfo` directly
- Provide meaningful error messages
- Include context in error messages
- Use `fmt.Errorf` with `%w` for error wrapping

### Logging

- Use structured logging with `log.Field`
- Include request IDs for tracing
- Log at appropriate levels (Debug, Info, Warn, Error)
- Never log sensitive information

## 🧪 Testing Guidelines

### Test Requirements

- **Coverage**: Minimum 90% for core modules
- **Types**: Unit tests, integration tests, and benchmarks
- **Naming**: Use descriptive test names
- **Isolation**: Tests should be independent and parallelizable

### Test Structure

```go
func TestFunctionName_Scenario_ExpectedResult(t *testing.T) {
    // Arrange
    // Act
    // Assert
}
```

### Running Tests

```bash
# All tests
make test

# Specific package
go test ./pkg/core/...

# With coverage
make test-coverage

# Benchmarks
make benchmark
```

## 📚 Documentation

### Code Documentation

- All public APIs must have Go doc comments
- Use English for all comments and documentation
- Include usage examples in doc comments
- Document parameters, return values, and errors

### README Updates

- Update README.md for new features
- Include usage examples
- Update badges and links
- Keep the quick start section current

### API Documentation

- Document all public interfaces
- Include parameter descriptions
- Provide usage examples
- Update when APIs change

## 🚀 Release Process

### Versioning

We follow [Semantic Versioning](https://semver.org/):
- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

### Release Checklist

- [ ] All tests pass (`make check`)
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] Version bumped in go.mod
- [ ] Release notes prepared
- [ ] Tag created
- [ ] Release published

## 🐛 Bug Reports

When reporting bugs, please include:

1. **Environment**
   - Go version
   - Operating system
   - EggyByte Core version

2. **Reproduction Steps**
   - Clear, numbered steps
   - Minimal code example
   - Expected vs actual behavior

3. **Additional Context**
   - Error messages
   - Logs (sanitized)
   - Related issues

## 💡 Feature Requests

For feature requests, please include:

1. **Problem Description**
   - What problem does this solve?
   - Why is it important?

2. **Proposed Solution**
   - How should it work?
   - Any design considerations?

3. **Alternatives**
   - Other solutions considered
   - Workarounds currently used

## 📞 Getting Help

- 💬 **Discord**: [Join our community](https://discord.gg/eggybyte)
- 📧 **Email**: [support@eggybyte.com](mailto:support@eggybyte.com)
- 🐛 **Issues**: [GitHub Issues](https://github.com/eggybyte-technology/go-eggybyte-core/issues)
- 📖 **Docs**: [Documentation](https://docs.eggybyte.com)

## 🏆 Recognition

Contributors will be:
- Listed in the project README
- Mentioned in release notes
- Invited to the core team (for significant contributions)

## 📄 License

By contributing to EggyByte Core, you agree that your contributions will be licensed under the Apache License 2.0.

---

Thank you for contributing to EggyByte Core! 🥚