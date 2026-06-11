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
- Go 1.23 or later
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

### 1. Install Go

**macOS:**
```bash
brew install go
```

**Linux:**
```bash
# Download from https://go.dev/dl/
wget https://go.dev/dl/go1.23.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.23.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

**Verify:**
```bash
go version
# Should output: go version go1.23.0 ...
```

### 2. Install Development Tools

**Automated (Recommended):**
```bash
make setup-dev
```

This installs:
- reuse (license compliance)
- pre-commit (git hooks)

**Manual Installation:**

**golangci-lint:**
```bash
# macOS
brew install golangci-lint

# Linux
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
```

**pre-commit:**
```bash
# Using pipx (recommended)
pipx install pre-commit

# Or using pip
pip install pre-commit
```

**reuse:**
```bash
# Using pipx (recommended)
pipx install reuse

# Or using pip
pip install reuse
```

**act (optional, for local CI):**
```bash
# macOS
brew install act

# Linux
curl https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash
```

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

### Repository Layout

```
legendary-funicular/
├── .github/              # GitHub Actions workflows
│   └── workflows/
│       ├── Release.yaml         # Release automation
│       ├── PRBuild.yaml         # PR validation
│       ├── golangci-lint.yaml   # Linting CI
│       ├── REUSE.yaml           # License compliance
│       └── scorecard.yml        # Security scanning
│
├── query/                # Query CLI tool
│   ├── cmd/             # CLI commands
│   │   ├── sql/        # SQL query command
│   │   ├── report/     # Report commands (list, describe, run)
│   │   ├── inspect/    # Inspect commands (dates, schema, config)
│   │   ├── dump/       # Dump command (debugging)
│   │   ├── version/    # Version command
│   │   ├── root.go     # Root command
│   │   └── utils.go    # Shared utilities (BuildCfg, etc.)
│   │
│   ├── internal/        # Internal packages (not exported)
│   │   ├── config/     # Configuration management
│   │   ├── sql/        # DuckDB integration
│   │   ├── render/     # Output formatting (JSON, NDJSON)
│   │   ├── report/     # Report system
│   │   ├── utils/      # Utilities (GetEnv, CheckFatal)
│   │   └── version/    # Version information
│   │
│   ├── go.mod          # Go module definition
│   ├── go.sum          # Go module checksums
│   ├── main.go         # Entry point
│   └── Dockerfile      # Container image
│
├── compactor/           # Compaction service
│   ├── internal/
│   │   ├── record/     # Log parsing
│   │   │   ├── syslog/      # Syslog parser (RFC3164/RFC5424)
│   │   │   └── cloudevent/  # CloudEvent parser (v1.0)
│   │   ├── pipeline/   # Streaming pipeline utilities
│   │   └── zio/        # Compression (zstd)
│   │
│   ├── go.mod          # Go module definition
│   ├── go.sum          # Go module checksums
│   ├── main.go         # Entry point
│   ├── compaction.go   # Main compaction logic
│   ├── util.go         # Utilities
│   └── Dockerfile      # Container image
│
├── collector/           # Log collection (Vector/FluentBit configs)
│   ├── vector.yaml
│   └── fluent-bit.conf
│
├── deploy/              # Deployment configurations
│   ├── systemd/        # Systemd service files
│   ├── kubernetes/     # Kubernetes manifests
│   └── docker-compose/ # Docker Compose files
│
├── docs/                # Documentation
│   ├── ARCHITECTURE.md # System architecture
│   ├── USER_GUIDE.md   # User guide
│   ├── DEVELOPMENT.md  # This file
│   ├── OPERATIONS.md   # Operations guide
│   └── API_REFERENCE.md # API reference
│
├── LICENSES/            # License files
│   └── MIT.txt
│
├── .goreleaser.yaml     # GoReleaser configuration
├── .pre-commit-config.yaml # Pre-commit hooks
├── Makefile             # Build automation
├── README.md            # Project README
├── REUSE.toml           # REUSE configuration
├── CONTRIBUTING.md      # Contribution guidelines
└── CHANGELOG.md         # Change log
```

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

**Why two modules?**
- Independent versioning
- Separate dependencies
- Cleaner builds

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

See [TESTING_QUICK_START.md](../TESTING_QUICK_START.md) for detailed testing guidelines.

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

## Development Workflow

### 1. Create a Branch

```bash
# Update main
git checkout main
git pull origin main

# Create feature branch
git checkout -b feature/my-feature
```

### 2. Make Changes

Edit code, add tests, update documentation.

### 3. Run Tests

```bash
# Run tests
make test

# Run linters
make lint

# Run all checks
make all
```

### 4. Commit Changes

```bash
# Stage changes
git add .

# Commit (pre-commit hooks run automatically)
git commit -m "Add feature X"
```

**Commit Message Format:**
```
<type>: <subject>

<body>

<footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `test`: Tests
- `refactor`: Code refactoring
- `chore`: Maintenance

**Example:**
```
feat: Add JSON field extraction to SQL queries

Adds support for json_extract_string() in SQL queries to extract
fields from the data column.

Closes #123
```

### 5. Push Changes

```bash
git push origin feature/my-feature
```

### 6. Create Pull Request

1. Go to GitHub
2. Click "New Pull Request"
3. Select your branch
4. Fill in description
5. Submit

### 7. Address Review Feedback

```bash
# Make changes
git add .
git commit -m "Address review feedback"
git push origin feature/my-feature
```

### 8. Merge

Once approved, maintainer will merge your PR.

---

## Contributing

### Contribution Guidelines

See [CONTRIBUTING.md](../CONTRIBUTING.md) for detailed guidelines.

**Quick summary:**
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Run linters
6. Submit a pull request

### Code Style

**Follow Go conventions:**
- Use `gofmt` (automatic via pre-commit)
- Use `golangci-lint` (automatic via pre-commit)
- Write godoc comments for exported functions
- Keep functions small and focused
- Use meaningful variable names

**Example:**
```go
// BuildCfg builds a configuration from environment variables and CLI flags.
// It returns an error if required configuration is missing.
func BuildCfg() (*Config, error) {
    endpoint := os.Getenv("S3_ENDPOINT")
    if endpoint == "" {
        return nil, fmt.Errorf("S3_ENDPOINT is required")
    }

    return &Config{
        Endpoint: endpoint,
        // ...
    }, nil
}
```

### Testing Requirements

**All PRs must:**
- Include tests for new code
- Maintain or improve coverage
- Pass all existing tests
- Pass linters

**Test coverage expectations:**
- New features: 80%+ coverage
- Bug fixes: Add test that reproduces bug
- Refactoring: Maintain existing coverage

### Documentation Requirements

**Update documentation when:**
- Adding new features
- Changing CLI flags
- Changing configuration
- Changing behavior

**Documentation to update:**
- README.md (if user-facing)
- docs/USER_GUIDE.md (for new commands/flags)
- docs/API_REFERENCE.md (for API changes)
- Code comments (godoc)

### Review Process

**PR checklist:**
- [ ] Tests added/updated
- [ ] Tests passing
- [ ] Linters passing
- [ ] Documentation updated
- [ ] Commit messages clear
- [ ] No merge conflicts

**Review criteria:**
- Code quality
- Test coverage
- Documentation
- Performance impact
- Security implications

---

## Debugging

### Debug with Delve

```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug query
cd query
dlv debug . -- sql "SELECT * FROM SOURCES LIMIT 10"

# Debug compactor
cd compactor
dlv debug .
```

### Debug with Print Statements

```go
import "log/slog"

slog.Debug("Debug message", "key", value)
slog.Info("Info message", "key", value)
slog.Warn("Warning message", "key", value)
slog.Error("Error message", "key", value)
```

### Debug with Environment Variables

```bash
# Enable DuckDB logging
export DUCKDB_LOG_LEVEL=DEBUG

# Enable S3 debug logging
export AWS_SDK_LOAD_CONFIG=1
export AWS_LOG_LEVEL=debug

# Run with debug logging
cd query
go run . sql "SELECT * FROM SOURCES LIMIT 10"
```

### Debug S3 Access

```bash
# Test S3 access with AWS CLI
aws s3 ls s3://openchami-logs-daily \
  --endpoint-url=$S3_ENDPOINT

# Test with curl
curl -v $S3_ENDPOINT
```

### Debug DuckDB Queries

```bash
# Use DuckDB CLI directly
duckdb <<SQL
CREATE SECRET local_s3 (
  TYPE s3,
  PROVIDER config,
  KEY_ID '$S3_ACCESS_KEY',
  SECRET '$S3_SECRET_KEY',
  REGION '$S3_REGION',
  ENDPOINT '$S3_ENDPOINT',
  USE_SSL false
);

SELECT * FROM read_parquet('s3://bucket/logs/*.parquet') LIMIT 10;
SQL
```

### Profiling

**CPU Profiling:**
```bash
cd query
go test -cpuprofile=cpu.prof -bench=. ./...
go tool pprof cpu.prof
```

**Memory Profiling:**
```bash
cd query
go test -memprofile=mem.prof -bench=. ./...
go tool pprof mem.prof
```

**Trace:**
```bash
cd query
go test -trace=trace.out -bench=. ./...
go tool trace trace.out
```

### Common Issues

**Issue: Tests fail with "S3 access denied"**

Solution: Set test environment variables:
```bash
export S3_ENDPOINT="http://localhost:7070"
export S3_ACCESS_KEY="test-key"
export S3_SECRET_KEY="test-secret"
```

**Issue: Linter fails with "module not found"**

Solution: Run go mod tidy:
```bash
cd query && go mod tidy
cd compactor && go mod tidy
```

**Issue: Pre-commit hooks fail**

Solution: Run hooks manually to see error:
```bash
pre-commit run --all-files
```

---

## Release Process

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
- **[API Reference](API_REFERENCE.md)** - CLI reference
- **[Testing Guide](../TESTING_QUICK_START.md)** - Testing methodology

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
