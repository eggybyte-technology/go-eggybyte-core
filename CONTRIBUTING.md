# Contributing to EggyByte Core

Thank you for your interest in contributing to EggyByte Core! This document provides guidelines and information for contributors.

## 📋 Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Contributing Guidelines](#contributing-guidelines)
- [Code Standards](#code-standards)
- [Testing Requirements](#testing-requirements)
- [Documentation Standards](#documentation-standards)
- [Pull Request Process](#pull-request-process)
- [Release Process](#release-process)

## 🤝 Code of Conduct

This project follows the [Contributor Covenant Code of Conduct](https://www.contributor-covenant.org/version/2/1/code_of_conduct/). By participating, you agree to uphold this code.

## 🚀 Getting Started

### Prerequisites

- Go 1.25.1 or later
- Git
- Make (for running development commands)

### Fork and Clone

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/your-username/go-eggybyte-core.git
   cd go-eggybyte-core
   ```

## 🛠️ Development Setup

### Install Dependencies

```bash
make deps
```

### Run Tests

```bash
make test
```

### Run Linting

```bash
make lint
```

### Generate Coverage Report

```bash
make test-coverage
```

## 📝 Contributing Guidelines

### Types of Contributions

We welcome contributions in the following areas:

- **Bug Fixes**: Fix issues reported in GitHub Issues
- **Feature Enhancements**: Add new functionality to existing packages
- **Documentation**: Improve README, code comments, and examples
- **Performance Improvements**: Optimize existing code
- **Test Coverage**: Add tests for uncovered code paths

### Before You Start

1. **Check Existing Issues**: Look for existing issues or discussions about your idea
2. **Create an Issue**: For significant changes, create an issue first to discuss the approach
3. **Small Changes**: For small fixes, you can submit a PR directly

## 🎯 Code Standards

### Go Code Style

- Follow standard Go formatting (`gofmt`)
- Use `golangci-lint` for code quality checks
- All public functions, types, and packages must have comprehensive English documentation
- Use meaningful variable and function names
- Keep functions focused and small (≤50 lines for public functions)

### Documentation Requirements

All public APIs must include:

```go
// PackageName provides functionality for...
package packagename

// TypeName represents a... 
// It provides methods for...
//
// Thread Safety: Safe for concurrent use when properly initialized
//
// Example:
//   instance := NewTypeName()
//   result := instance.DoSomething()
type TypeName struct {
    // FieldName represents...
    FieldName string
}

// MethodName performs a specific operation.
// It handles... and returns...
//
// Parameters:
//   - param1: Description of parameter
//   - param2: Description of parameter
//
// Returns:
//   - result: Description of return value
//   - error: Description of error conditions
//
// Example:
//   result, err := instance.MethodName("value1", "value2")
//   if err != nil {
//       return err
//   }
func (t *TypeName) MethodName(param1, param2 string) (result string, error) {
    // Implementation
}
```

### Error Handling

- Use wrapped errors with context: `fmt.Errorf("operation failed: %w", err)`
- Return meaningful error messages
- Use custom error types for different error categories
- Always handle errors explicitly

### Concurrency

- Use `sync.RWMutex` for read-write locks
- Use `atomic` operations for simple counters and flags
- Document thread safety guarantees
- Avoid data races (use `go test -race`)

## 🧪 Testing Requirements

### Test Coverage

- Maintain minimum 80% test coverage for new code
- Aim for 90%+ coverage for critical packages
- Use table-driven tests for multiple scenarios
- Include both positive and negative test cases

### Test Structure

```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name     string
        input    InputType
        expected ExpectedType
        wantErr  bool
    }{
        {
            name:     "successful_case",
            input:    validInput,
            expected: expectedOutput,
            wantErr:  false,
        },
        {
            name:     "error_case",
            input:    invalidInput,
            expected: nil,
            wantErr:  true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := FunctionName(tt.input)
            
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

### Running Tests

```bash
# Run all tests
make test

# Run tests with race detection
go test -race ./...

# Run tests with coverage
make test-coverage

# Run specific package tests
go test ./pkg/config
```

## 📚 Documentation Standards

### Code Comments

- All public functions, types, and packages must have English documentation
- Use complete sentences ending with periods
- Include examples for complex functions
- Document thread safety guarantees
- Explain parameters, return values, and error conditions

### README Updates

- Update README.md for new features or significant changes
- Keep examples current and working
- Update badges and links as needed

### API Documentation

- Use godoc-compatible comments
- Include usage examples
- Document all exported symbols

## 🔄 Pull Request Process

### Before Submitting

1. **Run Tests**: Ensure all tests pass
2. **Check Coverage**: Verify test coverage meets requirements
3. **Run Linting**: Fix any linting issues
4. **Update Documentation**: Update relevant documentation
5. **Rebase**: Rebase on latest main branch

### PR Description Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Tests pass locally
- [ ] New tests added for new functionality
- [ ] Test coverage maintained/improved

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] No breaking changes (or clearly documented)

## Related Issues
Fixes #(issue number)
```

### Review Process

1. **Automated Checks**: CI/CD pipeline runs tests and linting
2. **Code Review**: At least one maintainer reviews the PR
3. **Testing**: Reviewer tests the changes locally
4. **Approval**: PR approved and merged

## 🚀 Release Process

### Versioning

We follow [Semantic Versioning](https://semver.org/):
- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

### Release Steps

1. **Update Version**: Update version in go.mod and documentation
2. **Update CHANGELOG**: Add entry for new version
3. **Create Tag**: Create git tag for the version
4. **Publish**: Release is automatically published via GitHub Actions

### Changelog Format

```markdown
## [1.2.0] - 2024-01-15

### Added
- New feature description
- Another new feature

### Changed
- Changed behavior description

### Fixed
- Bug fix description

### Removed
- Removed feature description
```

## 🐛 Reporting Issues

### Bug Reports

When reporting bugs, please include:

1. **Environment**: Go version, OS, etc.
2. **Steps to Reproduce**: Clear steps to reproduce the issue
3. **Expected Behavior**: What you expected to happen
4. **Actual Behavior**: What actually happened
5. **Code Sample**: Minimal code that reproduces the issue
6. **Error Messages**: Full error messages and stack traces

### Feature Requests

For feature requests, please include:

1. **Use Case**: Why this feature would be useful
2. **Proposed Solution**: How you envision it working
3. **Alternatives**: Other solutions you've considered
4. **Additional Context**: Any other relevant information

## 📞 Getting Help

- **GitHub Issues**: For bugs and feature requests
- **GitHub Discussions**: For questions and general discussion
- **Documentation**: Check the README and code comments first

## 🎉 Recognition

Contributors will be recognized in:
- GitHub contributors list
- Release notes for significant contributions
- Project documentation

Thank you for contributing to EggyByte Core! 🥚
