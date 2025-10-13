# Contributing to EggyByte Core

Thank you for your interest in contributing to EggyByte Core! This document provides guidelines and information for contributors.

## Code of Conduct

By participating in this project, you agree to abide by our Code of Conduct. Please read and follow it in all your interactions with the project.

## Getting Started

### Prerequisites

- Go 1.25.1 or later
- Git
- Make (optional, for build automation)
- Docker (optional, for local development)

### Development Setup

1. **Fork the Repository**
   ```bash
   # Fork on GitHub, then clone your fork
   git clone https://github.com/your-username/go-eggybyte-core.git
   cd go-eggybyte-core
   ```

2. **Install Dependencies**
   ```bash
   go mod download
   ```

3. **Install CLI Tool**
   ```bash
   go install ./cmd/ebcctl
   ```

4. **Run Tests**
   ```bash
   make test
   ```

5. **Build Project**
   ```bash
   make build
   ```

## Development Workflow

### 1. Create a Feature Branch

```bash
git checkout -b feature/your-feature-name
```

### 2. Make Your Changes

Follow the coding standards and guidelines outlined below.

### 3. Test Your Changes

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run linters
make lint

# Run all checks
make check
```

### 4. Commit Your Changes

```bash
git add .
git commit -m "feat: add new feature"
```

### 5. Push and Create Pull Request

```bash
git push origin feature/your-feature-name
```

Then create a pull request on GitHub.

## Coding Standards

### Go Code Style

1. **Follow Go Conventions**
   - Use `gofmt` for formatting
   - Follow Go naming conventions
   - Use meaningful variable and function names

2. **Method Length**
   - Public methods: ≤50 lines
   - Private methods: ≤100 lines (recommended ≤50)
   - Test methods: ≤200 lines

3. **Error Handling**
   - Always handle errors explicitly
   - Use structured error messages
   - Return meaningful errors

4. **Context Usage**
   - Pass context as first parameter
   - Use context for cancellation and timeouts
   - Don't store context in structs

### Documentation Standards

1. **English Only**
   - All comments, documentation, and error messages must be in English
   - No Chinese or other languages allowed

2. **Comment Coverage**
   - Public APIs: 100% comment coverage
   - Private functions: ≥80% comment coverage
   - Complex logic: Must be commented

3. **Comment Format**
   ```go
   // FunctionName performs a specific operation with detailed description.
   // This function handles the complete workflow including validation,
   // processing, and error handling.
   //
   // Parameters:
   //   - ctx: Request context for cancellation and tracing
   //   - input: Input data for processing
   //
   // Returns:
   //   - *Result: Processed result data
   //   - error: Detailed error information if processing fails
   //
   // Example:
   //   result, err := FunctionName(ctx, input)
   //   if err != nil {
   //       log.Error("Processing failed", log.Field{Key: "error", Value: err})
   //   }
   func FunctionName(ctx context.Context, input *Input) (*Result, error) {
       // Implementation
   }
   ```

### Testing Standards

1. **Test Coverage**
   - Core module: ≥90% coverage
   - New features: 100% coverage
   - Bug fixes: Include regression tests

2. **Test Structure**
   ```go
   func TestFunctionName(t *testing.T) {
       tests := []struct {
           name     string
           input    Input
           expected Result
           wantErr  bool
       }{
           {
               name:     "successful_case",
               input:    Input{Value: "test"},
               expected: Result{Value: "test"},
               wantErr:  false,
           },
           {
               name:     "error_case",
               input:    Input{Value: ""},
               expected: Result{},
               wantErr:  true,
           },
       }

       for _, tt := range tests {
           t.Run(tt.name, func(t *testing.T) {
               result, err := FunctionName(context.Background(), tt.input)
               if tt.wantErr {
                   assert.Error(t, err)
               } else {
                   assert.NoError(t, err)
                   assert.Equal(t, tt.expected, result)
               }
           })
       }
   }
   ```

3. **Test Naming**
   - Use descriptive test names
   - Follow pattern: `TestFunctionName_Scenario`
   - Use table-driven tests for multiple scenarios

### Logging Standards

1. **Use Structured Logging**
   ```go
   // Good
   log.Info("User created",
       log.Field{Key: "user_id", Value: userID},
       log.Field{Key: "email", Value: email},
   )

   // Bad
   log.Printf("User created: %s", email)
   ```

2. **Log Levels**
   - `debug`: Detailed information for debugging
   - `info`: General information about operations
   - `warn`: Warning messages for potential issues
   - `error`: Error messages for failures
   - `fatal`: Fatal errors that cause program exit

3. **Don't Log Sensitive Data**
   - Never log passwords, tokens, or personal information
   - Use placeholders for sensitive fields
   - Implement data masking for logs

## Project Structure

### Core Modules

- `pkg/core/` - Bootstrap orchestrator and service lifecycle
- `pkg/config/` - Configuration management
- `pkg/log/` - Structured logging
- `pkg/db/` - Database abstraction
- `pkg/cache/` - Cache abstraction
- `pkg/service/` - Service lifecycle management
- `pkg/monitoring/` - Monitoring and health checks

### CLI Tool

- `cmd/ebcctl/` - Command-line tool for code generation
- `cmd/ebcctl/commands/` - Command implementations
- `cmd/ebcctl/templates/` - Code generation templates

### Documentation

- `docs/` - Project documentation
- `README.md` - Project overview
- `CONTRIBUTING.md` - This file
- `CHANGELOG.md` - Version history

## Pull Request Process

### Before Submitting

1. **Check Your Changes**
   ```bash
   make check  # Runs all checks: test, lint, vet, fmt
   ```

2. **Update Documentation**
   - Update README.md if needed
   - Add/update API documentation
   - Update examples if applicable

3. **Update Tests**
   - Add tests for new features
   - Update existing tests if needed
   - Ensure all tests pass

### Pull Request Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## Testing
- [ ] Added tests for new functionality
- [ ] All existing tests pass
- [ ] Test coverage maintained or improved

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] No sensitive data in logs
- [ ] Error handling implemented
- [ ] Context used appropriately
```

### Review Process

1. **Automated Checks**
   - CI/CD pipeline runs automatically
   - Tests must pass
   - Linters must pass
   - Coverage must meet requirements

2. **Manual Review**
   - Code review by maintainers
   - Architecture review for significant changes
   - Documentation review

3. **Approval**
   - At least one maintainer approval required
   - All checks must pass
   - No outstanding discussions

## Issue Reporting

### Bug Reports

When reporting bugs, please include:

1. **Environment Information**
   - Go version
   - Operating system
   - EggyByte Core version

2. **Reproduction Steps**
   - Clear steps to reproduce the issue
   - Expected behavior
   - Actual behavior

3. **Error Messages**
   - Full error messages
   - Stack traces if available
   - Log output

4. **Code Example**
   - Minimal code example that reproduces the issue
   - Configuration used

### Feature Requests

When requesting features, please include:

1. **Use Case**
   - Describe the problem you're trying to solve
   - Explain why existing functionality doesn't meet your needs

2. **Proposed Solution**
   - Describe your proposed solution
   - Include API design if applicable

3. **Alternatives Considered**
   - What alternatives have you considered?
   - Why is this approach better?

## Release Process

### Versioning

We follow [Semantic Versioning](https://semver.org/):

- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

### Release Steps

1. **Update Version**
   - Update version in `go.mod`
   - Update version in CLI tool
   - Update documentation

2. **Create Release**
   - Create git tag
   - Generate release notes
   - Publish to GitHub

3. **Update Documentation**
   - Update README.md
   - Update API documentation
   - Update examples

## Community Guidelines

### Communication

1. **Be Respectful**
   - Treat everyone with respect
   - Be constructive in feedback
   - Assume good intentions

2. **Be Inclusive**
   - Welcome newcomers
   - Help others learn
   - Share knowledge

3. **Be Professional**
   - Use professional language
   - Stay on topic
   - Be patient

### Getting Help

1. **Documentation**
   - Check the [docs directory](docs/)
   - Read the [API Reference](docs/API_REFERENCE.md)
   - Review [examples](docs/examples/)

2. **Community**
   - Use [GitHub Discussions](https://github.com/eggybyte-technology/go-eggybyte-core/discussions)
   - Ask questions in issues
   - Join community forums

3. **Support**
   - Email: support@eggybyte.com
   - GitHub Issues for bugs
   - GitHub Discussions for questions

## Recognition

Contributors will be recognized in:

1. **Contributors List**
   - Listed in README.md
   - Included in release notes

2. **Documentation**
   - Credit in documentation
   - Mention in changelog

3. **Community**
   - Recognition in community forums
   - Featured in project updates

## License

By contributing to EggyByte Core, you agree that your contributions will be licensed under the same license as the project.

## Questions?

If you have questions about contributing:

- **Documentation**: Check the [docs directory](docs/)
- **Issues**: Create a [GitHub Issue](https://github.com/eggybyte-technology/go-eggybyte-core/issues)
- **Discussions**: Use [GitHub Discussions](https://github.com/eggybyte-technology/go-eggybyte-core/discussions)
- **Email**: Contact us at support@eggybyte.com

Thank you for contributing to EggyByte Core! 🎉