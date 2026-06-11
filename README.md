<!--
SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC

SPDX-License-Identifier: MIT
-->

# openchami-logq

[![Go Version](https://img.shields.io/badge/go-1.23+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![REUSE status](https://api.reuse.software/badge/github.com/OpenCHAMI/legendary-funicular)](https://api.reuse.software/info/github.com/OpenCHAMI/legendary-funicular)

> **Active Development** — No stable release yet. Core functionality is stable
> but APIs may change. See [CHANGELOG.md](CHANGELOG.md) for breaking changes.

**Lightweight log lake + DuckDB query tool for OpenCHAMI.**

Store all your HPC cluster logs and events cheaply in S3-compatible storage, then query them instantly with SQL. No complex setup, no expensive infrastructure, no data loss.

## Why openchami-logq?

**The Problem:**
- Traditional log systems (ELK, Splunk) are expensive and complex
- Streaming platforms (Kafka, ClickHouse) require constant maintenance
- You need to keep logs for compliance but rarely query old data
- Schema changes break ingestion pipelines

**Our Solution:**
- **Cheap Storage:** NDJSON in S3 (pennies per GB)
- **Fast Queries:** DuckDB queries Parquet directly from S3
- **Schema Flexible:** Accept any log format, extract fields best-effort
- **Zero Maintenance:** No cluster to manage, no indices to maintain
- **HPC-Aware:** Built-in queries for xnames, services, traces

## Quick Start

### Prerequisites

- Go 1.23+ (for building from source)
- S3-compatible storage (VersityGW, MinIO, or AWS S3)
- Optional: Docker for containerized deployment

### Installation

**Option 1: Install Script (Recommended)**
```bash
curl -sSL https://raw.githubusercontent.com/OpenCHAMI/legendary-funicular/main/installer.bash | bash
```

**Option 2: Build from Source**
```bash
git clone https://github.com/OpenCHAMI/legendary-funicular.git
cd legendary-funicular
make build
sudo cp bin/openchami-logq-* /usr/local/bin/
```

**Option 3: Docker**
```bash
docker pull ghcr.io/openchami/logq-query:latest
docker pull ghcr.io/openchami/logq-compactor:latest
```

**Option 4: RPM Package** (Coming Soon)
```bash
# RPM packaging in progress
```

### Quick Configuration

Create environment file `~/.openchami-logq.env`:

```bash
# S3 Configuration
export S3_ENDPOINT="http://localhost:7070"
export S3_REGION="us-east-1"
export S3_ACCESS_KEY="your-access-key"
export S3_SECRET_KEY="your-secret-key"
export S3_BUCKET_RAW="openchami-logs-raw"
export S3_BUCKET_COMPACTED="openchami-logs-daily"
export S3_SSL="false"  # true for HTTPS
```

Source the environment:
```bash
source ~/.openchami-logq.env
```

### Your First Query

**Query recent errors:**
```bash
openchami-logq-query sql \
  --scope compacted \
  --stream logs \
  "SELECT ts, host, level, msg
   FROM SOURCES
   WHERE level = 'ERROR'
   ORDER BY ts DESC
   LIMIT 10"
```

**Use a built-in report:**
```bash
# List available reports
openchami-logq-query report list

# Run a report
openchami-logq-query report run find-all-service-errors
```

**Inspect your data:**
```bash
# See available dates
openchami-logq-query inspect dates --stream logs

# View schema
openchami-logq-query inspect schema --stream logs

# Check configuration
openchami-logq-query inspect config
```

## Architecture

```
┌─────────────┐
│   Syslog    │──┐
│  Collector  │  │
└─────────────┘  │
                 ├──> ┌──────────────┐      ┌──────────────┐
┌─────────────┐  │    │              │      │              │
│ CloudEvents │──┼───>│  S3 Storage  │─────>│  Compactor   │
│  Collector  │  │    │   (NDJSON)   │      │  (Parquet)   │
└─────────────┘  │    └──────────────┘      └──────────────┘
                 │           │                      │
┌─────────────┐  │           │                      │
│   Vector/   │──┘           v                      v
│  FluentBit  │         ┌──────────────────────────────┐
└─────────────┘         │       Query Engine           │
                        │   (DuckDB + S3 Direct)       │
                        └──────────────────────────────┘
                                    │
                                    v
                            ┌──────────────┐
                            │  CLI Output  │
                            │ JSON/NDJSON  │
                            └──────────────┘
```

### Components

1. **Collector** (Vector/FluentBit)
   - Accepts syslog and CloudEvents
   - Writes NDJSON to S3 raw bucket
   - No schema validation (never drops logs)

2. **Compactor** (`openchami-logq-compactor`)
   - Runs daily (or on-demand)
   - Converts NDJSON → Parquet
   - Extracts fields best-effort
   - Writes to compacted bucket

3. **Query** (`openchami-logq-query`)
   - CLI tool for querying logs
   - Uses DuckDB to query Parquet from S3
   - Supports custom SQL and built-in reports
   - Outputs JSON or NDJSON

## Key Features

### Schema Flexibility

**Problem:** Traditional log systems break when log format changes.

**Solution:** We store raw NDJSON (never loses data) and extract fields best-effort during compaction.

```bash
# Syslog format
{"ts": "2026-06-10T12:00:00Z", "host": "node01", "msg": "Started"}

# CloudEvent format
{"specversion": "1.0", "type": "system.event", "source": "/compute/node01"}

# Both work! Fields extracted where possible.
```

### Cost Optimization

**Storage Costs:**
- Raw NDJSON: ~$0.023/GB/month (S3 Standard)
- Compacted Parquet: ~$0.004/GB/month (10x compression + S3 IA)
- **Example:** 1TB logs/month = ~$50/month (vs $1000+ for traditional)

**Query Costs:**
- DuckDB queries Parquet directly from S3
- Only pay for data scanned (not stored)
- **Example:** Query 1GB = ~$0.0004 (vs $5+ for indexed search)

### Built-in Reports

Pre-built queries for common tasks:

```bash
# Find all parse errors
openchami-logq-query report run find-all-parse-errors

# Find service errors
openchami-logq-query report run find-all-service-errors

# List all reports
openchami-logq-query report list

# Describe a report
openchami-logq-query report describe find-all-service-errors
```

### HPC-Specific Features

**xname Support** (Coming Soon):
```sql
SELECT * FROM SOURCES
WHERE xname LIKE 'x3000c0s%'
ORDER BY ts DESC
```

**Service Tracing** (Coming Soon):
```sql
SELECT * FROM SOURCES
WHERE trace_id = 'abc123'
ORDER BY ts ASC
```

## Usage Examples

### Basic Queries

**Count logs by level:**
```bash
openchami-logq-query sql \
  "SELECT level, COUNT(*) as count
   FROM SOURCES
   GROUP BY level
   ORDER BY count DESC"
```

**Find logs from specific host:**
```bash
openchami-logq-query sql \
  --stream logs \
  "SELECT * FROM SOURCES
   WHERE host = 'node01'
   AND ts > '2026-06-10'"
```

**Query CloudEvents:**
```bash
openchami-logq-query sql \
  --stream events \
  "SELECT type, source, COUNT(*) as count
   FROM SOURCES
   GROUP BY type, source"
```

### Advanced Queries

**Time-range analysis:**
```bash
openchami-logq-query sql \
  "SELECT
     DATE_TRUNC('hour', ts) as hour,
     level,
     COUNT(*) as count
   FROM SOURCES
   WHERE ts BETWEEN '2026-06-10' AND '2026-06-11'
   GROUP BY hour, level
   ORDER BY hour, level"
```

**JSON field extraction:**
```bash
openchami-logq-query sql \
  "SELECT
     json_extract_string(data, '$.user') as user,
     COUNT(*) as count
   FROM SOURCES
   WHERE level = 'ERROR'
   GROUP BY user"
```

**Cross-stream query:**
```bash
# Query both logs and events
openchami-logq-query sql \
  --stream logs,events \
  "SELECT type, COUNT(*) FROM SOURCES GROUP BY type"
```

### Output Formats

**JSON (default):**
```bash
openchami-logq-query sql "SELECT * FROM SOURCES LIMIT 2"
# Output: [{"ts":"...","host":"..."},{"ts":"...","host":"..."}]
```

**NDJSON (streaming):**
```bash
openchami-logq-query sql --format ndjson "SELECT * FROM SOURCES LIMIT 2"
# Output: {"ts":"...","host":"..."}
#         {"ts":"...","host":"..."}
```

**Pipe to jq:**
```bash
openchami-logq-query sql "SELECT * FROM SOURCES LIMIT 10" | jq '.[] | select(.level=="ERROR")'
```

## Compaction

The compactor runs periodically to convert raw NDJSON to compressed Parquet:

```bash
# Run compactor manually
openchami-logq-compactor

# Run for specific date
openchami-logq-compactor --date 2026-06-10

# Dry run (no changes)
openchami-logq-compactor --dry-run
```

**Typical Schedule:**
```cron
# Run daily at 2 AM
0 2 * * * /usr/local/bin/openchami-logq-compactor
```

**What it does:**
1. Lists NDJSON files in raw bucket for date
2. Streams and parses each file (syslog or cloudevent)
3. Converts to Parquet with compression
4. Uploads to compacted bucket
5. Deletes raw files (after successful upload)

## Configuration

### Environment Variables

**S3 Configuration:**
```bash
S3_ENDPOINT       # S3 endpoint URL (required)
S3_REGION         # S3 region (default: us-east-1)
S3_ACCESS_KEY     # S3 access key (required)
S3_SECRET_KEY     # S3 secret key (required)
S3_BUCKET_RAW     # Raw NDJSON bucket (default: openchami-logs-raw)
S3_BUCKET_COMPACTED # Compacted Parquet bucket (default: openchami-logs-daily)
S3_SSL            # Use SSL/TLS (default: true)
```

**Query Configuration:**
```bash
LOGQ_FORMAT       # Output format: json|ndjson (default: json)
LOGQ_SCOPE        # Data scope: raw|compacted (default: compacted)
LOGQ_STREAM       # Log stream: logs|events (default: logs)
```

**Compactor Configuration:**
```bash
COMPACTOR_DATE    # Date to compact (default: yesterday)
COMPACTOR_DRY_RUN # Dry run mode (default: false)
```

### CLI Flags

All environment variables can be overridden with CLI flags:

```bash
openchami-logq-query sql \
  --endpoint "http://localhost:7070" \
  --access-key "..." \
  --secret-key "..." \
  --scope compacted \
  --stream logs \
  --format ndjson \
  "SELECT * FROM SOURCES LIMIT 10"
```

See `openchami-logq-query --help` for all options.

## Development

### Prerequisites

- Go 1.23+
- golangci-lint
- pre-commit (optional)
- Docker (optional)

### Setup Development Environment

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

### Project Structure

```
legendary-funicular/
├── query/              # Query CLI tool
│   ├── cmd/           # CLI commands
│   │   ├── sql/       # SQL query command
│   │   ├── report/    # Report commands
│   │   ├── inspect/   # Inspect commands
│   │   └── version/   # Version command
│   └── internal/      # Internal packages
│       ├── config/    # Configuration
│       ├── sql/       # DuckDB integration
│       ├── render/    # Output formatting
│       └── report/    # Report system
├── compactor/         # Compaction service
│   ├── internal/
│   │   ├── record/    # Log parsing (syslog, cloudevent)
│   │   ├── pipeline/  # Streaming pipeline
│   │   └── zio/       # Compression (zstd)
│   └── main.go
├── collector/         # Log collection (Vector config)
├── deploy/            # Deployment configs
├── .github/           # CI/CD workflows
└── docs/              # Documentation
```

### Testing

```bash
# Run all tests
make test

# Run specific tests
make test-query
make test-compactor

# Run tests with coverage
make test-coverage

# Run tests with race detector
go test -race ./...

# Run benchmarks
go test -bench=. ./...
```

### Code Quality

```bash
# Format code
make fmt

# Run linters
make lint

# Fix linter issues
make lint-fix

# Check for vulnerabilities
make vuln

# REUSE compliance
make reuse
```

## Deployment

### Systemd Services

Example systemd service for compactor:

```ini
[Unit]
Description=OpenCHAMI Log Compactor
After=network.target

[Service]
Type=simple
User=logq
EnvironmentFile=/etc/openchami-logq/env
ExecStart=/usr/local/bin/openchami-logq-compactor
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

### Docker Compose

```yaml
version: '3.8'
services:
  compactor:
    image: ghcr.io/openchami/logq-compactor:latest
    environment:
      - S3_ENDPOINT=http://versitygw:7070
      - S3_ACCESS_KEY=${S3_ACCESS_KEY}
      - S3_SECRET_KEY=${S3_SECRET_KEY}
    restart: unless-stopped

  query:
    image: ghcr.io/openchami/logq-query:latest
    environment:
      - S3_ENDPOINT=http://versitygw:7070
      - S3_ACCESS_KEY=${S3_ACCESS_KEY}
      - S3_SECRET_KEY=${S3_SECRET_KEY}
    command: version
```

### Kubernetes

See [deploy/kubernetes/](deploy/kubernetes/) for Kubernetes manifests.

## Documentation

- **[Architecture](docs/ARCHITECTURE.md)** - System design and components
- **[User Guide](docs/USER_GUIDE.md)** - Detailed usage instructions
- **[Developer Guide](docs/DEVELOPMENT.md)** - Development setup and guidelines
- **[Operations Guide](docs/OPERATIONS.md)** - Deployment and operations
- **[API Reference](docs/API_REFERENCE.md)** - Configuration and CLI reference
- **[Testing Guide](TESTING_QUICK_START.md)** - Testing methodology

## Performance

**Query Performance:**
- Simple queries: <100ms
- Complex aggregations: <1s
- Full table scans: <5s (1GB data)

**Compaction Performance:**
- ~10MB/s throughput
- ~10:1 compression ratio
- ~1 hour for 1TB of logs

**Storage Efficiency:**
- Raw NDJSON: 1.0x
- Compressed NDJSON (zstd): 0.3x
- Parquet: 0.1x (10x compression)

## Troubleshooting

**Query fails with "S3 access denied":**
```bash
# Check credentials
aws s3 ls s3://openchami-logs-daily --endpoint-url=$S3_ENDPOINT

# Verify environment
openchami-logq-query inspect config
```

**Compactor fails:**
```bash
# Run with verbose logging
openchami-logq-compactor --verbose

# Check S3 permissions
# Compactor needs: s3:GetObject, s3:PutObject, s3:DeleteObject
```

**Query returns no results:**
```bash
# Check available dates
openchami-logq-query inspect dates --stream logs

# Verify data exists
aws s3 ls s3://openchami-logs-daily/logs/ --endpoint-url=$S3_ENDPOINT
```

See [docs/TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md) for more.

## Contributing

We welcome contributions! Please see:

- **[CONTRIBUTING.md](CONTRIBUTING.md)** - Contribution guidelines
- **[CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)** - Community guidelines
- **[Development Guide](docs/DEVELOPMENT.md)** - Development setup

### Quick Contribution Guide

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests (`make test`)
5. Run linters (`make lint`)
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

## Testing

This project has comprehensive test coverage:

- **240 test cases** covering all major components
- **51 benchmarks** for performance validation
- **2 fuzz tests** for parser robustness
- **~50% code coverage** (100% on critical paths)
- **2 bugs found and fixed** during testing

See [TESTING_QUICK_START.md](TESTING_QUICK_START.md) for testing guidelines.

## Roadmap

- [x] Phase 0: Storage & IAM
- [x] Phase 1: Syslog ingestion
- [x] Phase 2: CloudEvents support
- [x] Phase 3: Compaction
- [x] Phase 4: Query CLI
- [x] Phase 5: Built-in reports
- [ ] Phase 6: Web UI (planned)
- [ ] Phase 7: Real-time queries (planned)
- [ ] Phase 8: Alert system (planned)

## License

MIT License - see [LICENSE](LICENSE) for details.

Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC

## Acknowledgments

- **OpenCHAMI Community** - For feedback and testing
- **DuckDB** - Amazing embedded analytics database
- **VersityGW** - S3-compatible storage gateway
- **Contributors** - See [CONTRIBUTORS.md](CONTRIBUTORS.md)

## Support

- **Issues:** [GitHub Issues](https://github.com/OpenCHAMI/legendary-funicular/issues)
- **Discussions:** [GitHub Discussions](https://github.com/OpenCHAMI/legendary-funicular/discussions)
- **Slack:** [OpenCHAMI Slack](https://openchami.slack.com)
- **Email:** support@openchami.org

## Related Projects

- **[OpenCHAMI](https://github.com/OpenCHAMI)** - Open Composable Heterogeneous Adaptable Management Infrastructure
- **[metadata-service](https://github.com/OpenCHAMI/metadata-service)** - OpenCHAMI metadata service
- **[DuckDB](https://duckdb.org)** - In-process SQL OLAP database
- **[VersityGW](https://github.com/versity/versitygw)** - S3-compatible storage gateway

---

**Made with ❤️ by the OpenCHAMI community**
