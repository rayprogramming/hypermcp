# Contributing to hypermcp

Thank you for your interest in contributing to hypermcp! This document provides guidelines and instructions for contributing.

## Table of Contents
- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Pull Request Process](#pull-request-process)
- [Coding Standards](#coding-standards)
- [Testing Guidelines](#testing-guidelines)
- [Documentation](#documentation)
- [Commit Messages](#commit-messages)

## Code of Conduct

This project follows the [Contributor Covenant Code of Conduct](https://www.contributor-covenant.org/version/2/1/code_of_conduct/). By participating, you are expected to uphold this code. Please report unacceptable behavior to the project maintainers.

## Getting Started

### Prerequisites
- Go 1.22 or later
- Git
- Make (optional, but recommended)

### Fork and Clone
1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/hypermcp.git
   cd hypermcp
   ```
3. Add the upstream repository:
   ```bash
   git remote add upstream https://github.com/rayprogramming/hypermcp.git
   ```

### Install Dependencies
```bash
make mod-download
```

### Install Development Tools
```bash
make install-tools
```

## Development Workflow

### 1. Create a Branch
```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/your-bug-fix
```

### 2. Make Your Changes
- Write your code
- Add or update tests
- Update documentation

### 3. Run Tests and Checks
```bash
# Run all checks
make check

# Or run individual checks
make test           # Run tests
make test-race      # Run tests with race detector
make lint           # Run linter
make fmt            # Format code
make vet            # Run go vet
```

### 4. Commit Your Changes
Follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:
```bash
git add .
git commit -m "feat: add new cache eviction strategy"
```

### 5. Keep Your Branch Updated
```bash
git fetch upstream
git rebase upstream/main
```

### 6. Push Your Changes
```bash
git push origin feature/your-feature-name
```

## Pull Request Process

### Before Submitting
- [ ] All tests pass (`make test`)
- [ ] Code is formatted (`make fmt`)
- [ ] Linter passes (`make lint`)
- [ ] Documentation is updated
- [ ] Commit messages follow conventions
- [ ] Branch is up to date with main

### Submitting a PR
1. Go to your fork on GitHub
2. Click "New Pull Request"
3. Select your branch
4. Fill out the PR template with:
   - Clear description of changes
   - Related issue numbers (if any)
   - Breaking changes (if any)
   - Screenshots (if UI changes)

### PR Review Process
- Maintainers will review your PR within 1-2 weeks
- Address any requested changes
- Once approved, maintainers will merge your PR

### After Merge
- Delete your feature branch
- Pull the latest changes from upstream

## Coding Standards

### Go Style Guide
- Follow [Effective Go](https://go.dev/doc/effective_go)
- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` for formatting
- Keep functions focused and concise

### Package Structure
```
hypermcp/
├── cache/          # Cache implementation
├── httpx/          # HTTP client
├── *.go            # Core server implementation
└── *_test.go       # Tests
```

### Error Handling
- Always handle errors explicitly
- Wrap errors with context: `fmt.Errorf("context: %w", err)`
- Use custom error types for domain-specific errors

### Example Good Code
```go
// Good: Clear function name, error handling, documentation
func (s *Server) AddTool[In, Out any](tool *mcp.Tool, handler mcp.ToolHandlerFor[In, Out]) error {
    if tool == nil {
        return fmt.Errorf("tool cannot be nil")
    }
    
    mcp.AddTool(s.mcp, tool, handler)
    s.IncrementToolCount()
    return nil
}
```

## Testing Guidelines

### Test Coverage
- Aim for >80% test coverage
- All new code must include tests
- Test both success and failure cases

### Writing Tests
```go
func TestServerNew(t *testing.T) {
    tests := []struct {
        name    string
        config  Config
        wantErr bool
    }{
        {
            name: "valid config",
            config: Config{
                Name:    "test",
                Version: "1.0.0",
            },
            wantErr: false,
        },
        // More test cases...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Running Tests
```bash
# All tests
make test

# With coverage
make test-coverage

# With race detector
make test-race

# Specific package
go test ./cache -v
```

### Benchmarks
```bash
# Run benchmarks
make bench

# Specific benchmark
go test ./cache -bench=BenchmarkCache_Get -benchmem
```

## Documentation

### Code Documentation
- All exported types, functions, and methods must have godoc comments
- Comments should explain **why**, not **what**
- Include usage examples for complex functionality

### Example godoc Comment
```go
// AddTool registers a tool with the MCP server and automatically increments the tool counter.
//
// This is a generic function that provides type-safe tool registration. The input and output
// types are inferred from the handler function signature.
//
// Example:
//
//	hypermcp.AddTool(srv, &mcp.Tool{Name: "echo"}, func(...) { ... })
func AddTool[In, Out any](s *Server, tool *mcp.Tool, handler mcp.ToolHandlerFor[In, Out]) {
    // Implementation
}
```

### README and Guides
- Keep README.md up to date
- Update EXAMPLE.md for new features
- Add examples for complex features

## Commit Messages

We follow [Conventional Commits](https://www.conventionalcommits.org/) specification.

### Format
```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting)
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `test`: Adding or updating tests
- `chore`: Maintenance tasks
- `ci`: CI/CD changes

### Examples
```bash
# Feature
feat: add cache eviction strategy

# Bug fix
fix: resolve race condition in cache cleanup

# Documentation
docs: update README with new examples

# Breaking change
feat!: change Config struct fields

BREAKING CHANGE: Config.CacheTTL is now Config.CacheConfig.TTL
```

### Scope (Optional)
- `cache`: Changes to cache package
- `httpx`: Changes to HTTP client
- `server`: Changes to server implementation
- `deps`: Dependency updates

## Questions?

- Open an issue for questions
- Join discussions on GitHub
- Check existing issues and PRs

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to hypermcp! 🎉
