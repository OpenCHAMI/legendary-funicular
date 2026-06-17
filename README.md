<!--
SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC

SPDX-License-Identifier: MIT
-->

# openchami-logq

[![Go Version](https://img.shields.io/badge/go-1.26+-blue.svg)](https://golang.org)
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

### Installation

**Recommended:** Use the install script
```bash
curl -sSL https://raw.githubusercontent.com/OpenCHAMI/legendary-funicular/main/installer.bash | bash
```

**Other options:** See [Installation Options](#installation-options) below for Docker, building from source, or RPM packages.

### Configuration

Set your S3 connection details:
```bash
export S3_ENDPOINT="http://localhost:7070"
export S3_ACCESS_KEY="your-access-key"
export S3_SECRET_KEY="your-secret-key"
```

For all configuration options, see [Configuration](#configuration) below or run `openchami-logq-query --help`.

### Your First Query

```bash
# Query error logs
openchami-logq-query sql "SELECT level, COUNT(*) FROM SOURCES GROUP BY level"

# Use a built-in report
openchami-logq-query report list
openchami-logq-query report run find-all-service-errors

# Inspect your data
openchami-logq-query inspect dates --stream logs
```

That's it! See [Usage Examples](#usage-examples) for more.

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

## Installation Options

### Install Script (Recommended)
```bash
curl -sSL https://raw.githubusercontent.com/OpenCHAMI/legendary-funicular/main/installer.bash | bash
```

### Docker
```bash
docker pull ghcr.io/openchami/logq-query:latest
docker pull ghcr.io/openchami/logq-compactor:latest
```

### Build from Source
```bash
git clone https://github.com/OpenCHAMI/legendary-funicular.git
cd legendary-funicular
make build
sudo cp bin/openchami-logq-* /usr/local/bin/
```

**Prerequisites:** Go 1.26+, S3-compatible storage (VersityGW, MinIO, or AWS S3)

### RPM Package (Coming Soon)
```bash
# RPM packaging in progress
```

For development setup and testing, see [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md).

## Usage Examples

**Query logs by level:**
```bash
openchami-logq-query sql "SELECT level, COUNT(*) FROM SOURCES GROUP BY level ORDER BY COUNT(*) DESC"
```

**Find logs from specific host:**
```bash
openchami-logq-query sql --stream logs "SELECT * FROM SOURCES WHERE host = 'node01' AND ts > '2026-06-10'"
```

**Query CloudEvents:**
```bash
openchami-logq-query sql --stream events "SELECT type, source, COUNT(*) FROM SOURCES GROUP BY type, source"
```

**Time-range analysis:**
```bash
openchami-logq-query sql "SELECT DATE_TRUNC('hour', ts) as hour, level, COUNT(*) as count FROM SOURCES WHERE ts BETWEEN '2026-06-10' AND '2026-06-11' GROUP BY hour, level"
```

**Output formats:**
```bash
# JSON (default)
openchami-logq-query sql "SELECT * FROM SOURCES LIMIT 5"

# NDJSON for streaming
openchami-logq-query sql --format ndjson "SELECT * FROM SOURCES LIMIT 5"

# Pipe to jq
openchami-logq-query sql "SELECT * FROM SOURCES LIMIT 10" | jq '.[] | select(.level=="ERROR")'
```

**For more advanced examples** including JSON field extraction, window functions, and DuckDB-specific optimizations, see [docs/USER_GUIDE.md](docs/USER_GUIDE.md).

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

openchami-logq is configured via environment variables or CLI flags. CLI flags override environment variables.

### S3 Connection

The tools need to connect to S3-compatible storage:
- **Endpoint:** URL of your S3 service (required)
- **Credentials:** Access key and secret key for authentication (required)
- **Buckets:** Separate buckets for raw logs and compacted data (defaults: `openchami-logs-raw`, `openchami-logs-daily`)
- **Region:** S3 region (default: `us-east-1`)
- **SSL:** Enable/disable HTTPS (default: `true`)

### Query Options

Control query behavior:
- **Format:** Output format - `json` for complete results or `ndjson` for streaming (default: `json`)
- **Scope:** Data source - `raw` (NDJSON), `compacted` (Parquet), or `all` (default: `compacted`)
- **Stream:** Log type - `logs` or `events` (default: `logs`)

### Example Configuration

```bash
# Minimal setup
export S3_ENDPOINT="http://localhost:7070"
export S3_ACCESS_KEY="your-access-key"
export S3_SECRET_KEY="your-secret-key"

# Optional overrides
export S3_BUCKET_RAW="my-raw-logs"
export S3_BUCKET_COMPACTED="my-compacted-logs"
export S3_SSL="false"  # for local development
```

**All options:** Run `openchami-logq-query --help` for complete reference.

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

For detailed operational procedures, monitoring, and production deployment, see [docs/OPERATIONS.md](docs/OPERATIONS.md).

## Development & Testing

This project has comprehensive test coverage including unit tests, benchmarks, and fuzz tests.
For development setup, testing guidelines, and contribution workflow, see [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md).

Quick start for contributors:
```bash
git clone https://github.com/OpenCHAMI/legendary-funicular.git
cd legendary-funicular
make setup-dev  # Install development tools
make build      # Build binaries
make test       # Run test suite
make lint       # Run linters
```

## Documentation

- **[User Guide](docs/USER_GUIDE.md)** - Advanced usage, SQL examples, and DuckDB tips
- **[Operations Guide](docs/OPERATIONS.md)** - Production deployment and operational procedures
- **[Architecture](docs/ARCHITECTURE.md)** - System design and technical deep dive
- **[Developer Guide](docs/DEVELOPMENT.md)** - Development setup and testing guidelines
- **[Changelog](CHANGELOG.md)** - Version history and breaking changes

## Performance

Performance varies based on hardware, data size, query complexity, and S3 network latency.

**Query Performance:**
- DuckDB's columnar engine enables fast aggregations on compressed Parquet files
- Columnar filtering scans only required columns, reducing I/O
- Query time depends on data scanned, predicate selectivity, and network bandwidth
- Typical queries on modest hardware: simple aggregations complete in seconds

**Compaction Performance:**
- Throughput depends on CPU cores, network bandwidth to S3, and compression settings
- Streaming architecture processes data incrementally without loading full files
- Typical compression ratios: 5:1 to 15:1 depending on log structure and field cardinality

**Storage Efficiency:**
- Raw NDJSON: baseline (1.0x)
- Compressed NDJSON (zstd): ~70% reduction (0.3x)
- Parquet (columnar + compression): ~85-95% reduction (0.05x-0.15x)

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

For additional troubleshooting, check the project's GitHub issues or open a new issue with details about your environment and error messages.

## Contributing

We welcome contributions! See [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) for development setup and testing guidelines.

For organization-wide guidelines, see:
- [OpenCHAMI Contributing Guidelines](https://github.com/OpenCHAMI/.github/blob/main/CONTRIBUTING.md)
- [OpenCHAMI Code of Conduct](https://github.com/OpenCHAMI/.github/blob/main/CODE_OF_CONDUCT.md)

**Quick start:**
1. Fork the repository and create a feature branch
2. Make your changes and add tests
3. Run `make test` and `make lint`
4. Submit a pull request

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
