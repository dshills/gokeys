# Contributing to gokeys

Thank you for your interest in contributing to gokeys! This document provides guidelines and instructions for contributing.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)
- [Coding Standards](#coding-standards)
- [Platform Testing](#platform-testing)

## Code of Conduct

Be respectful, professional, and constructive in all interactions.

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR_USERNAME/gokeys.git`
3. Add upstream remote: `git remote add upstream https://github.com/dshills/gokeys.git`
4. Create a feature branch: `git checkout -b feature/your-feature-name`

## Development Setup

### Prerequisites

- Go 1.21 or higher
- golangci-lint for linting
- git

### Install Dependencies

```bash
# Install linter
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Download project dependencies
go mod download
```

### Build

```bash
go build ./...
```

## Testing

We have comprehensive CI/CD that tests on multiple platforms and Go versions.

### Run Tests Locally

```bash
# Run all tests
go test ./...

# Run with race detector
go test -race ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test -v ./input

# Run benchmarks
go test -bench=. -benchmem ./input
```

### Test Guidelines

- All new features must have tests
- Maintain or improve code coverage (target: >80%)
- Tests must pass on Linux, macOS, and Windows
- Use table-driven tests where appropriate
- Include benchmarks for performance-critical code

### Platform-Specific Testing

Our CI automatically tests on:
- **Linux**: Ubuntu 20.04, 22.04, latest
- **macOS**: macOS 12, 13, latest
- **Windows**: Windows 2019, 2022, latest
- **Go versions**: 1.21, 1.22, 1.23

If you can test locally on different platforms, that's great! Otherwise, the CI will validate.

## Linting

Code must pass all linters:

```bash
golangci-lint run
```

Our linting rules include:
- gosec (security)
- gocritic (code quality)
- staticcheck
- govet
- errcheck (all errors must be checked)
- cyclop (max complexity 30)
- revive

## Submitting Changes

### Commit Message Format

Use clear, descriptive commit messages:

```
type: brief description (50 chars or less)

More detailed explanation if needed (wrap at 72 characters).
Include motivation for the change and contrast with previous behavior.

- Bullet points are okay
- Use present tense ("add feature" not "added feature")
- Reference issues: Fixes #123, Closes #456

Co-authored-by: Name <email@example.com>
```

**Types**: `feat`, `fix`, `docs`, `test`, `refactor`, `perf`, `chore`, `ci`

### Pull Request Process

1. **Update your branch**
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Run full test suite**
   ```bash
   go test -race -cover ./...
   golangci-lint run
   ```

3. **Push to your fork**
   ```bash
   git push origin feature/your-feature-name
   ```

4. **Create Pull Request**
   - Use a clear, descriptive title
   - Fill out the PR template
   - Link related issues
   - Add screenshots/examples if applicable

5. **CI Checks**
   - All tests must pass on all platforms
   - Linting must pass
   - Code coverage should not decrease

6. **Code Review**
   - Address review feedback
   - Keep discussions professional and constructive
   - Update PR based on feedback

### PR Checklist

- [ ] Tests added/updated
- [ ] Documentation updated
- [ ] Benchmarks added for performance changes
- [ ] golangci-lint passes
- [ ] All tests pass locally
- [ ] Commit messages follow format
- [ ] PR description is clear

## Coding Standards

### Go Code Style

Follow standard Go conventions:
- Use `gofmt` for formatting (automatic in most editors)
- Follow [Effective Go](https://go.dev/doc/effective_go)
- Use meaningful variable names
- Keep functions focused and small
- Comment exported types and functions

### Specific Rules

1. **Error Handling**
   ```go
   // Good
   if err := doSomething(); err != nil {
       return fmt.Errorf("failed to do something: %w", err)
   }

   // Bad - ignored error
   doSomething()
   ```

2. **Error Messages**
   - Start with lowercase
   - No trailing punctuation
   - Provide context
   ```go
   return fmt.Errorf("failed to parse key %v: %w", key, err)
   ```

3. **Comments**
   ```go
   // Good - explains why
   // Use RWMutex to allow concurrent reads while preventing race conditions

   // Bad - explains what (obvious from code)
   // Lock the mutex
   ```

4. **Exported Names**
   All exported functions/types MUST have godoc comments:
   ```go
   // GameInput provides action mapping for game development.
   // It wraps the Input interface with logical action names.
   type GameInput interface {
       // ...
   }
   ```

### Performance Considerations

- Minimize allocations in hot paths
- Use sync.Pool for frequently allocated objects
- Benchmark performance-critical code
- Profile before optimizing

### Thread Safety

- Document thread-safety guarantees
- Use appropriate synchronization primitives
- Test with race detector
- Consider lock granularity

## Platform Testing

### What We Need

We especially welcome testing and bug reports from:

- **Linux**: Various distributions (Ubuntu, Debian, Arch, Fedora, etc.)
- **Windows**: Windows 10, 11, Server editions
- **Terminal Emulators**:
  - Linux: gnome-terminal, konsole, xterm, alacritty, kitty
  - macOS: iTerm2, Terminal.app, Alacritty, Kitty
  - Windows: Windows Terminal, ConEmu, Cmder, mintty

### How to Help with Platform Testing

1. Run the test suite on your platform
2. Try the examples
3. Report any failures with:
   - Platform/OS version
   - Terminal emulator and version
   - Go version
   - Error messages and stack traces
   - Steps to reproduce

### Terminal-Specific Issues

If you encounter terminal-specific behavior:
- Document which terminal it affects
- Include escape sequences if relevant
- Test on multiple terminals if possible

## Documentation

### Update Documentation When

- Adding new features
- Changing behavior
- Fixing bugs that affect usage
- Adding examples

### Documentation Files

- `README.md` - Main documentation
- `CLAUDE.md` - Development notes
- Package godoc comments
- Example code in `examples/`

## Questions?

- Open an issue for bugs
- Start a discussion for questions
- Check existing issues/discussions first

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

**Thank you for contributing to gokeys!** 🎉
