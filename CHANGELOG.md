<!--
SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC

SPDX-License-Identifier: MIT
-->

# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Comprehensive unit tests for compactor and query packages
- Benchmark tests for record parsing and pipeline operations
- Fuzz tests for cloudevent and syslog parsing
- E2E tests for compaction and query workflows
- Test utilities for S3 mocking and test data generation

### Changed
- **BREAKING**: Go version requirement updated to 1.26+
- Updated all dependencies to latest versions
- Improved documentation accuracy by removing unsourced performance claims
- Consolidated documentation to reduce duplication

### Fixed
- Pipeline construction logic to properly handle transform errors
- Documentation references to non-existent files removed

## [0.1.0] - 2026-06-10

### Added
- Initial release of openchami-logq
- Compactor service for NDJSON to Parquet conversion
  - Support for syslog format parsing
  - Support for CloudEvents format parsing
  - Zstd compression support
  - Streaming pipeline architecture
- Query CLI with DuckDB integration
  - SQL query command with S3 direct access
  - Built-in report system
  - Multiple output formats (JSON, NDJSON)
  - Inspect commands (dates, schema, config)
- S3-compatible storage support
  - Works with VersityGW, MinIO, AWS S3
  - Configurable via environment variables or CLI flags
- Project infrastructure
  - Makefile with common development tasks
  - GitHub Actions CI/CD workflows
  - golangci-lint configuration
  - REUSE compliance for licensing
  - Docker support

### Architecture
- Collector layer (Vector/FluentBit integration)
- Storage layer (S3-compatible object storage)
- Compaction layer (NDJSON → Parquet conversion)
- Query layer (DuckDB with S3 direct read)

[Unreleased]: https://github.com/OpenCHAMI/legendary-funicular/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/OpenCHAMI/legendary-funicular/releases/tag/v0.1.0
