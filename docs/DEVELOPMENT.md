<!--
SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC

SPDX-License-Identifier: MIT
-->

# Development Guide

**Project:** openchami-logq
**Version:** 1.0
**Last Updated:** June 10, 2026

---

## Table of Contents

1. [Getting Started](#getting-started)
2. [Development Environment Setup](#development-environment-setup)
3. [Project Structure](#project-structure)
4. [Building](#building)
5. [Testing](#testing)
6. [Code Quality](#code-quality)
7. [Development Workflow](#development-workflow)
8. [Contributing](#contributing)
9. [Debugging](#debugging)
10. [Release Process](#release-process)

---

## Getting Started

### Prerequisites

Before you begin development, ensure you have:

**Required:**
- Go 1.26 or later
- Git
- Make

**Recommended:**
- golangci-lint (for linting)
- pre-commit (for git hooks)
- reuse (for license compliance)
- Docker (for container builds)
- act (for local CI testing)

### Quick Start

```bash
# Clone repository
git clone https://github.com/OpenCHAMI/legendary-funicular.git
cd legendary-funicular

# Install development tools
make setup-dev

# Build binaries
make build

# Run tests
make test

# Run linters
make lint
```

---

## Development Environment Setup

### 1. Install Go 1.26 or greater

See the main [go website](https://go.dev/) for installation details.

### 2. Install Development Tools

**Automated (Recommended):**
```bash
make setup-dev
```

This installs:
- reuse (license compliance)
- pre-commit (git hooks)


### 3. Install Pre-commit Hooks

```bash
pre-commit install
pre-commit install --hook-type commit-msg
```

This ensures code quality checks run before every commit.

### 4. Verify Setup

```bash
# Check Go
go version

# Check golangci-lint
golangci-lint version

# Check pre-commit
pre-commit --version

# Check reuse
reuse --version
```

---

## Project Structure

### Key Directories

**`query/`** - Query CLI tool
- Main user-facing command
- Uses DuckDB for SQL queries
- Supports multiple output formats

**`compactor/`** - Compaction service
- Converts NDJSON → Parquet
- Runs daily (cron job)
- Streaming pipeline (constant memory)

**`collector/`** - Log collection configs
- Vector and FluentBit configurations
- Not Go code (just configs)

**`internal/`** - Internal packages
- Not exported (Go convention)
- Implementation details

### Module Structure

This project uses Go modules with two separate modules:

```
legendary-funicular/
├── query/go.mod       # Query module
└── compactor/go.mod   # Compactor module
```

### Package Naming Conventions

- **cmd/** - CLI commands (one per file)
- **internal/** - Private packages (not exported)
- **pkg/** - Public packages (none currently)

---

## Building

### Build Everything

```bash
make build
```

This builds both binaries:
- `bin/openchami-logq-query`
- `bin/openchami-logq-compactor`

### Build Specific Component

```bash
# Query only
make build-query

# Compactor only
make build-compactor
```

### Build with Version Information

```bash
# Automatic version from git
make build

# Or specify version
VERSION=v1.0.0 make build
```

Version information is embedded via ldflags:
```go
var (
    version = "dev"
    commit  = "unknown"
    date    = "unknown"
)
```

### Build for Different Platforms

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 make build

# Linux ARM64
GOOS=linux GOARCH=arm64 make build

# macOS ARM64 (M1/M2)
GOOS=darwin GOARCH=arm64 make build
```

### Docker Build

```bash
# Build both images
make docker-build

# Build specific image
make docker-build-query
make docker-build-compactor
```

### GoReleaser Build

```bash
# Snapshot release (no tag required)
make release-snapshot

# This creates:
# - Multi-arch binaries
# - Docker images
# - Archives (tar.gz, zip)
# - Checksums
# - SBOM
```

---

## Testing

### Test Coverage

This project has comprehensive test coverage:
- **240 test cases** covering all major components
- **51 benchmarks** for performance validation
- **2 fuzz tests** for parser robustness
- **~50% code coverage** with 100% coverage on critical paths

### Run All Tests

```bash
make test
```

This runs tests for both modules with:
- Race detector (`-race`)
- Coverage (`-coverprofile`)
- Atomic coverage mode (`-covermode=atomic`)

### Run Specific Tests

```bash
# Query tests only
make test-query

# Compactor tests only
make test-compactor
```

### Run Single Test

```bash
# Run specific test
cd query
go test -v -run TestBuildCfg ./cmd

# Run specific test in package
cd compactor
go test -v -run TestSyslogParser ./internal/record/syslog
```

### Coverage Reports

```bash
# Generate HTML coverage reports
make test-coverage

# Open in browser
open coverage-query.html
open coverage-compactor.html
```

### Benchmarks

```bash
# Run all benchmarks
cd query
go test -bench=. ./...

cd compactor
go test -bench=. ./...

# Run specific benchmark
cd compactor
go test -bench=BenchmarkSyslogParse ./internal/record/syslog
```

### Fuzzing

```bash
# Fuzz syslog parser
cd compactor
go test -fuzz=FuzzSyslogParse ./internal/record/syslog

# Fuzz cloudevent parser
go test -fuzz=FuzzCloudEventParse ./internal/record/cloudevent
```

### Test with Race Detector

```bash
# Race detector (slower but catches concurrency bugs)
cd query
go test -race ./...

cd compactor
go test -race ./...
```

### Test Verbose Output

```bash
# Show all test output
go test -v ./...

# Show only failures
go test ./...
```

### Testing Guidelines

**Key principles:**
1. Use table-driven tests
2. Use testify/assert for readability
3. Test error paths
4. Test edge cases
5. Add benchmarks for performance-critical code

---

## Code Quality

### Linting

**Run linters:**
```bash
# Lint everything
make lint

# Lint specific module
make lint-query
make lint-compactor

# Auto-fix issues
make lint-fix
```

### Formatting

```bash
# Format code
make fmt

# This runs:
# - go fmt
# - goimports (if installed)
```

### Vet

```bash
# Run go vet
make vet
```

### Vulnerability Scanning

```bash
# Check for known vulnerabilities
make vuln
```

### License Compliance

```bash
# Check REUSE compliance
make reuse

# Generate SPDX bill of materials
make reuse-spdx

# Add REUSE headers to new files
make reuse-annotate
```

### Pre-commit Hooks

Pre-commit hooks run automatically on `git commit`:

```bash
# Run manually
pre-commit run --all-files

# Update hooks
pre-commit autoupdate
```

Hooks include:
- gofmt
- go mod tidy
- go vet
- golangci-lint
- REUSE compliance
- trailing whitespace
- end-of-file fixer

### CI Checks

All CI checks can be run locally:

```bash
# Run all checks
make all

# This runs:
# 1. clean
# 2. install (go mod download)
# 3. lint
# 4. test
# 5. build
```

### Local CI Testing (with act)

```bash
# List workflows
make act-list

# Run lint workflow
make act-lint

# Run REUSE workflow
make act-reuse
```

---


## Contributing

### Contribution Guidelines

See the [OpenCHAMI Contributing Guidelines](https://github.com/OpenCHAMI/.github/blob/main/CONTRIBUTING.md) for detailed guidelines.





### Releases

### Versioning

We follow [Semantic Versioning](https://semver.org/):

- **MAJOR** version for incompatible API changes
- **MINOR** version for new functionality (backwards compatible)
- **PATCH** version for bug fixes (backwards compatible)

### Release Checklist

1. **Update CHANGELOG.md**
   ```markdown
   ## [1.0.0] - 2026-06-10

   ### Added
   - New feature X

   ### Changed
   - Changed behavior Y

   ### Fixed
   - Fixed bug Z
   ```

2. **Update version in code** (if applicable)

3. **Create and push tag**
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```

4. **GitHub Actions automatically:**
   - Runs tests
   - Runs linters
   - Builds binaries (multi-arch)
   - Builds Docker images
   - Creates GitHub release
   - Uploads artifacts
   - Generates SBOM
   - Signs artifacts

5. **Verify release**
   - Check GitHub Releases page
   - Test binaries
   - Test Docker images

### Snapshot Releases

For testing before official release:

```bash
make release-snapshot
```

This creates local builds in `dist/` without pushing to GitHub.

### Release Artifacts

Each release includes:
- **Binaries:** Linux (amd64, arm64), macOS (amd64, arm64), Windows (amd64)
- **Archives:** tar.gz, zip
- **Docker Images:** ghcr.io/openchami/logq-query, ghcr.io/openchami/logq-compactor
- **Checksums:** SHA256SUMS
- **SBOM:** Software Bill of Materials
- **Provenance:** SLSA attestations

---

## Additional Resources

### Documentation

- **[Architecture](ARCHITECTURE.md)** - System design
- **[User Guide](USER_GUIDE.md)** - Usage instructions
- **[Operations Guide](OPERATIONS.md)** - Deployment and operations
- **[Testing Guide](../TESTING_QUICK_START.md)** - Testing methodology
- **[Changelog](../CHANGELOG.md)** - Version history

### External Resources

- **[Go Documentation](https://go.dev/doc/)** - Go language
- **[DuckDB Documentation](https://duckdb.org/docs/)** - DuckDB SQL
- **[Cobra Documentation](https://cobra.dev/)** - CLI framework
- **[testify Documentation](https://pkg.go.dev/github.com/stretchr/testify)** - Testing framework

### Community

- **[GitHub Issues](https://github.com/OpenCHAMI/legendary-funicular/issues)** - Bug reports and feature requests
- **[GitHub Discussions](https://github.com/OpenCHAMI/legendary-funicular/discussions)** - Questions and discussions
- **[OpenCHAMI Slack](https://openchami.slack.com)** - Real-time chat

---

## Getting Help

**Questions?**
- Open a [GitHub Discussion](https://github.com/OpenCHAMI/legendary-funicular/discussions)
- Ask on [OpenCHAMI Slack](https://openchami.slack.com)

**Bug Reports?**
- Open a [GitHub Issue](https://github.com/OpenCHAMI/legendary-funicular/issues)

**Security Issues?**
- Email security@openchami.org (do not open public issue)

---

**Happy coding! 🚀**
