<!--
SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC

SPDX-License-Identifier: MIT
-->

# openchami-logq

[![Go Version](https://img.shields.io/badge/go-1.26+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![REUSE status](https://api.reuse.software/badge/github.com/OpenCHAMI/legendary-funicular)](https://api.reuse.software/info/github.com/OpenCHAMI/legendary-funicular)

> [!WARNING]
>
> **Active Development** — No stable release yet. Core functionality is stable
> but APIs may change.

**Lightweight log lake + DuckDB query tool for OpenCHAMI.**

Store all your HPC OpenCHAMI cluster logs and events cheaply in S3-compatible
storage, then query them instantly with SQL. No complex setup, no expensive
infrastructure, no data loss.

## Why openchami-logq?

**The Problem:**

- Traditional log systems (ELK, Splunk) are expensive and complex
- Streaming platforms (Kafka, ClickHouse) require constant maintenance
- You need to keep logs for compliance but rarely query old data
- Schema changes break ingestion pipelines

**Our Solution:**

- **Cheap Storage:** Parquet in S3 (pennies per GB)
- **Fast Queries:** DuckDB queries Parquet directly from S3
- **Schema Flexible:** Accept any log format, extract fields best-effort
- **Zero Maintenance:** No cluster to manage, no indices to maintain
- **HPC-Aware:** Built-in queries for xnames, services, traces

## Quick Start

### Installation

See [Installation Options](#installation-options).

### Configuration

#### Common

Set your S3 connection details:

```bash
export S3_ENDPOINT="http://localhost:7070"
export S3_REGION="us-east-1"
export S3_ACCESS_KEY="your-access-key"
export S3_SECRET_KEY="your-secret-key"
```

#### Query CLI

```bash
export S3_BUCKET_NDJSON="openchami-logs-raw"
export S3_BUCKET_PARQUET="openchami-logs-daily"
```
Additional configuration options are available in the form of CLI args. Run
`openchami-logq-query --help` for more details.

#### Compactor

```bash
export S3_BUCKET_SOURCE="openchami-logs-raw"
export S3_BUCKET_SINK="openchami-logs-daily"
```

### Your First Query

```bash
# Query error logs
openchami-logq sql -S compacted "SELECT level, COUNT(level) as occurrences FROM SOURCES GROUP BY level"

# Use a built-in report
openchami-logq-query report list
openchami-logq-query report run find-all-service-errors

# Inspect the schema
openchami-logq-query inspect schema --stream logs

# Inspect your data
openchami-logq-query inspect dates --stream logs
```

That's it! See [Usage Examples](#usage-examples) for more.

## Key Features

### Schema Flexibility

**Problem:** Traditional log systems break when log format changes.

**Solution:** We store raw NDJSON (never loses data) and extract fields
best-effort during compaction (structured+unstructured schema).

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

**xname Support**:

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
$ git clone https://github.com/OpenCHAMI/legendary-funicular.git
$ cd legendary-funicular && bash ./installer.bash
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

## Query CLI Usage Examples

**Query logs by level:**

```bash
openchami-logq-query sql "
SELECT
    level,
    COUNT(level) AS occurrences
FROM
    SOURCES
GROUP BY
    level
ORDER BY
    occurrences DESC;
"
```

**Find logs from specific host:**

```bash
openchami-logq-query sql --stream logs "
SELECT
    *
FROM
    SOURCES
WHERE
    host = 'srv00';
"
```

**Query CloudEvents:**

```bash
openchami-logq-query sql --stream events "SELECT * FROM SOURCES"
```

**Time-range analysis:**

```bash
openchami-logq-query sql "
SELECT
    DATE_TRUNC('hour', ts::TIMESTAMP) AS hour,
    level,
    COUNT(level) AS occurrences
FROM
    SOURCES
WHERE
    ts::TIMESTAMP BETWEEN '2026-04-01'
                      AND '2026-04-02'
GROUP BY
    hour,
    level
ORDER BY
    hour,
    level;
"
```

**Output formats:**

```bash
# NDJSON (default)
openchami-logq-query sql "SELECT * FROM SOURCES LIMIT 5"

# JSON
openchami-logq-query sql --f json "SELECT * FROM SOURCES LIMIT 5"

# Pipe to jq
openchami-logq-query sql "
SELECT
    *
FROM
    SOURCES
LIMIT
    10;
" | jq 'select(.level == "err")'
```

**For more advanced examples** including JSON field extraction, window
functions, and DuckDB-specific optimizations, see
[docs/USER_GUIDE.md](docs/USER_GUIDE.md).

## Compaction

The compactor runs periodically to convert raw NDJSON to compressed Parquet.

**What it does:**

1. Lists NDJSON files in raw bucket for date
2. Streams and parses each file (syslog or cloudevent)
3. Converts to Parquet with compression
4. Uploads to compacted bucket
5. Deletes raw files (after successful upload)

**Typical Schedule:**

```cron
# Run daily at 2 AM
0 2 * * * /usr/local/bin/openchami-logq-compactor
```

> [!NOTE]
>
> If using the installation defaults (recommended), the compactor executes
> periodically as a Systemd oneshot service.

To run the compactor manually:

```bash
# Run compactor manually
openchami-logq-compactor
```

## Documentation

- **[User Guide](docs/USER_GUIDE.md)** - Advanced usage, SQL examples, and
  DuckDB tips
- **[Architecture](docs/ARCHITECTURE.md)** - System design and technical deep
  dive
- **[Developer Guide](docs/DEVELOPMENT.md)** - Development setup and testing
  guidelines
- **[Troubleshooting Guide](docs/TROUBLESHOOTING.md)** - Tips for common issues
  and pitfalls

## License

MIT License - see [LICENSE](LICENSE) for details.

Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC

## Acknowledgments

- **OpenCHAMI Community** - For feedback and testing
- **DuckDB** - Amazing embedded analytics database
- **VersityGW** - S3-compatible storage gateway

## Support

- **Issues:**
  [GitHub Issues](https://github.com/OpenCHAMI/legendary-funicular/issues)
- **Discussions:**
  [GitHub Discussions](https://github.com/OpenCHAMI/legendary-funicular/discussions)
- **Slack:** [OpenCHAMI Slack](https://openchami.slack.com)
- **Email:** support@openchami.org

## Related Projects

- **[OpenCHAMI](https://github.com/OpenCHAMI)** - Open Composable Heterogeneous
  Adaptable Management Infrastructure
- **[DuckDB](https://duckdb.org)** - In-process SQL OLAP database
- **[VersityGW](https://github.com/versity/versitygw)** - S3-compatible storage
  gateway
