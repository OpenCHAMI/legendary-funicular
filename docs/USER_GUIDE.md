<!--
SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC

SPDX-License-Identifier: MIT
-->

# User Guide

**Project:** openchami-logq
**Version:** 1.0
**Last Updated:** June 10, 2026

---

## Table of Contents

1. [Getting Started](#getting-started)
2. [Installation](#installation)
3. [Configuration](#configuration)
4. [Basic Usage](#basic-usage)
5. [SQL Queries](#sql-queries)
6. [Reports](#reports)
7. [Inspection](#inspection)
8. [Output Formats](#output-formats)
9. [Advanced Usage](#advanced-usage)
10. [Troubleshooting](#troubleshooting)
11. [Best Practices](#best-practices)
12. [FAQ](#faq)

---

## Getting Started

### What is openchami-logq?

openchami-logq is a log lake query tool that lets you:

- **Store** all your logs cheaply in S3
- **Query** them with SQL using DuckDB
- **Analyze** HPC cluster logs without complex infrastructure

### Quick Example

```bash
# Query recent errors
openchami-logq-query sql \
  "SELECT ts, host, level, msg
   FROM SOURCES
   WHERE level = 'ERROR'
   ORDER BY ts DESC
   LIMIT 10"
```

### Prerequisites

Before you begin, you need:

1. **S3-compatible storage** (VersityGW, MinIO, or AWS S3)
2. **Access credentials** (access key + secret key)
3. **Data in S3** (from collector or sample data)

---

## Installation

### Option 1: Install Script (Recommended)

The easiest way to install:

```bash
curl -sSL https://raw.githubusercontent.com/OpenCHAMI/legendary-funicular/main/installer.bash | bash
```

This installs both binaries to `/usr/local/bin/`:
- `openchami-logq-query` - Query tool
- `openchami-logq-compactor` - Compaction service

### Option 2: Build from Source

If you have Go 1.23+ installed:

```bash
# Clone repository
git clone https://github.com/OpenCHAMI/legendary-funicular.git
cd legendary-funicular

# Build binaries
make build

# Install to system
sudo cp bin/openchami-logq-* /usr/local/bin/

# Verify installation
openchami-logq-query version
```

### Option 3: Docker

Run in a container:

```bash
# Pull image
docker pull ghcr.io/openchami/logq-query:latest

# Run query
docker run --rm \
  -e S3_ENDPOINT="http://host.docker.internal:7070" \
  -e S3_ACCESS_KEY="your-key" \
  -e S3_SECRET_KEY="your-secret" \
  ghcr.io/openchami/logq-query:latest \
  sql "SELECT COUNT(*) FROM SOURCES"
```

### Option 4: RPM Package (Coming Soon)

```bash
# RPM packaging in progress
sudo yum install openchami-logq
```

### Verify Installation

```bash
# Check version
openchami-logq-query version

# Check help
openchami-logq-query --help
```

---

## Configuration

### Environment Variables

Create `~/.openchami-logq.env`:

```bash
# S3 Configuration (Required)
export S3_ENDPOINT="http://localhost:7070"
export S3_REGION="us-east-1"
export S3_ACCESS_KEY="your-access-key"
export S3_SECRET_KEY="your-secret-key"

# Bucket Names (Optional - these are defaults)
export S3_BUCKET_RAW="openchami-logs-raw"
export S3_BUCKET_COMPACTED="openchami-logs-daily"

# SSL/TLS (Optional - default is true)
export S3_SSL="false"  # Set to true for HTTPS

# Query Defaults (Optional)
export LOGQ_FORMAT="json"        # or ndjson
export LOGQ_SCOPE="compacted"    # or raw or all
export LOGQ_STREAM="logs"        # or events
```

Load the configuration:

```bash
source ~/.openchami-logq.env
```

### CLI Flags Override

All environment variables can be overridden with CLI flags:

```bash
openchami-logq-query sql \
  --endpoint "http://other-server:7070" \
  --access-key "different-key" \
  --secret-key "different-secret" \
  --scope compacted \
  --stream logs \
  --format json \
  "SELECT * FROM SOURCES LIMIT 10"
```

### Configuration Priority

Configuration is loaded in this order (later overrides earlier):

1. Default values
2. Environment variables
3. CLI flags

### Verify Configuration

Check your configuration:

```bash
openchami-logq-query inspect config
```

Output:
```json
{
  "s3_endpoint": "http://localhost:7070",
  "s3_region": "us-east-1",
  "s3_bucket_raw": "openchami-logs-raw",
  "s3_bucket_compacted": "openchami-logs-daily",
  "s3_ssl": false,
  "format": "json",
  "scope": "compacted",
  "stream": "logs"
}
```

---

## Basic Usage

### Command Structure

```
openchami-logq-query [global-flags] <command> [command-flags] [args]
```

### Available Commands

| Command | Description |
|---------|-------------|
| `sql` | Execute SQL queries |
| `report` | Run built-in reports |
| `inspect` | Inspect data and configuration |
| `dump` | Dump raw data (debugging) |
| `version` | Show version information |

### Global Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--scope` | `-S` | Data scope (all, compacted, raw) | `compacted` |
| `--stream` | `-s` | Data stream (logs, events) | `logs` |
| `--format` | `-f` | Output format (json, ndjson) | `json` |
| `--output` | `-o` | Output file path | stdout |

### Your First Query

**Count all logs:**
```bash
openchami-logq-query sql "SELECT COUNT(*) as total FROM SOURCES"
```

**View recent logs:**
```bash
openchami-logq-query sql \
  "SELECT * FROM SOURCES
   ORDER BY ts DESC
   LIMIT 10"
```

**Filter by level:**
```bash
openchami-logq-query sql \
  "SELECT * FROM SOURCES
   WHERE level = 'ERROR'
   LIMIT 10"
```

---

## SQL Queries

### Basic SELECT

**All columns:**
```bash
openchami-logq-query sql "SELECT * FROM SOURCES LIMIT 10"
```

**Specific columns:**
```bash
openchami-logq-query sql \
  "SELECT ts, host, level, msg FROM SOURCES LIMIT 10"
```

### WHERE Clause

**Filter by level:**
```bash
openchami-logq-query sql \
  "SELECT * FROM SOURCES WHERE level = 'ERROR'"
```

**Filter by host:**
```bash
openchami-logq-query sql \
  "SELECT * FROM SOURCES WHERE host = 'node01'"
```

**Multiple conditions:**
```bash
openchami-logq-query sql \
  "SELECT * FROM SOURCES
   WHERE level = 'ERROR'
   AND host LIKE 'node%'"
```

### Time Ranges

**Specific date:**
```bash
openchami-logq-query sql \
  "SELECT * FROM SOURCES
   WHERE ts >= '2026-06-10'
   AND ts < '2026-06-11'"
```

**Last 24 hours:**
```bash
openchami-logq-query sql \
  "SELECT * FROM SOURCES
   WHERE ts > NOW() - INTERVAL 24 HOUR"
```

**Date range:**
```bash
openchami-logq-query sql \
  "SELECT * FROM SOURCES
   WHERE ts BETWEEN '2026-06-01' AND '2026-06-10'"
```

### Aggregations

**Count by level:**
```bash
openchami-logq-query sql \
  "SELECT level, COUNT(*) as count
   FROM SOURCES
   GROUP BY level
   ORDER BY count DESC"
```

**Count by host:**
```bash
openchami-logq-query sql \
  "SELECT host, COUNT(*) as count
   FROM SOURCES
   GROUP BY host
   ORDER BY count DESC"
```

**Hourly counts:**
```bash
openchami-logq-query sql \
  "SELECT
     DATE_TRUNC('hour', ts) as hour,
     COUNT(*) as count
   FROM SOURCES
   GROUP BY hour
   ORDER BY hour"
```

### Sorting

**Newest first:**
```bash
openchami-logq-query sql \
  "SELECT * FROM SOURCES ORDER BY ts DESC LIMIT 10"
```

**Oldest first:**
```bash
openchami-logq-query sql \
  "SELECT * FROM SOURCES ORDER BY ts ASC LIMIT 10"
```

**Multiple columns:**
```bash
openchami-logq-query sql \
  "SELECT * FROM SOURCES ORDER BY level, ts DESC"
```

### LIMIT and OFFSET

**First 10 rows:**
```bash
openchami-logq-query sql \
  "SELECT * FROM SOURCES LIMIT 10"
```

**Pagination:**
```bash
# Page 1 (rows 1-10)
openchami-logq-query sql \
  "SELECT * FROM SOURCES LIMIT 10 OFFSET 0"

# Page 2 (rows 11-20)
openchami-logq-query sql \
  "SELECT * FROM SOURCES LIMIT 10 OFFSET 10"

# Page 3 (rows 21-30)
openchami-logq-query sql \
  "SELECT * FROM SOURCES LIMIT 10 OFFSET 20"
```

### JSON Field Extraction

**Extract JSON field:**
```bash
openchami-logq-query sql \
  "SELECT
     json_extract_string(data, '$.user') as user,
     COUNT(*) as count
   FROM SOURCES
   WHERE data IS NOT NULL
   GROUP BY user"
```

**Extract nested field:**
```bash
openchami-logq-query sql \
  "SELECT
     json_extract_string(data, '$.request.method') as method,
     json_extract_string(data, '$.request.path') as path
   FROM SOURCES"
```

### Pattern Matching

**LIKE operator:**
```bash
openchami-logq-query sql \
  "SELECT * FROM SOURCES WHERE msg LIKE '%error%'"
```

**Case-insensitive:**
```bash
openchami-logq-query sql \
  "SELECT * FROM SOURCES WHERE LOWER(msg) LIKE '%error%'"
```

**Multiple patterns:**
```bash
openchami-logq-query sql \
  "SELECT * FROM SOURCES
   WHERE msg LIKE '%error%'
   OR msg LIKE '%fail%'"
```

### DISTINCT Values

**Unique hosts:**
```bash
openchami-logq-query sql \
  "SELECT DISTINCT host FROM SOURCES ORDER BY host"
```

**Unique levels:**
```bash
openchami-logq-query sql \
  "SELECT DISTINCT level FROM SOURCES ORDER BY level"
```

### Subqueries

**Top 10 hosts by error count:**
```bash
openchami-logq-query sql \
  "SELECT host, error_count
   FROM (
     SELECT host, COUNT(*) as error_count
     FROM SOURCES
     WHERE level = 'ERROR'
     GROUP BY host
   )
   ORDER BY error_count DESC
   LIMIT 10"
```

### SOURCES Placeholder

The `SOURCES` keyword is automatically replaced with the correct S3 paths based on your flags:

```bash
# Queries compacted logs
openchami-logq-query sql --scope compacted --stream logs \
  "SELECT * FROM SOURCES"
# Becomes: SELECT * FROM read_parquet('s3://bucket/logs/*.parquet')

# Queries raw NDJSON
openchami-logq-query sql --scope raw --stream logs \
  "SELECT * FROM SOURCES"
# Becomes: SELECT * FROM read_json('s3://bucket-raw/logs/**/*.ndjson.zst')

# Queries both logs and events
openchami-logq-query sql --stream logs,events \
  "SELECT * FROM SOURCES"
# Becomes: SELECT * FROM (...) UNION ALL BY NAME SELECT * FROM (...)
```

---

## Reports

Reports are pre-built queries for common tasks.

### List Reports

```bash
openchami-logq-query report list
```

Output:
```
Available reports:
  - find-all-parse-errors: Find all logs with parse errors
  - find-all-service-errors: Find all service errors grouped by service
  - count-by-level: Count logs by level
  - top-hosts: Top 10 hosts by log volume
```

### Describe Report

```bash
openchami-logq-query report describe find-all-service-errors
```

Output:
```
Report: find-all-service-errors
Description: Find all service errors grouped by service
Parameters: (none)
Query:
  SELECT
    service,
    COUNT(*) as error_count,
    MIN(ts) as first_seen,
    MAX(ts) as last_seen
  FROM SOURCES
  WHERE level = 'ERROR'
  AND service IS NOT NULL
  GROUP BY service
  ORDER BY error_count DESC
```

### Run Report

```bash
openchami-logq-query report run find-all-service-errors
```

### Report with Parameters (Future)

```bash
# Future: Reports with parameters
openchami-logq-query report run errors-by-host --param host=node01
```

### Built-in Reports

#### find-all-parse-errors

Find logs that failed to parse:

```bash
openchami-logq-query report run find-all-parse-errors
```

**Use case:** Identify malformed logs that need investigation.

#### find-all-service-errors

Find errors grouped by service:

```bash
openchami-logq-query report run find-all-service-errors
```

**Use case:** Identify which services have the most errors.

#### count-by-level

Count logs by level:

```bash
openchami-logq-query report run count-by-level
```

**Use case:** Quick overview of log distribution.

#### top-hosts

Top 10 hosts by log volume:

```bash
openchami-logq-query report run top-hosts
```

**Use case:** Identify noisy hosts.

---

## Inspection

Inspection commands help you understand your data.

### Inspect Dates

See which dates have data:

```bash
openchami-logq-query inspect dates
```

Output:
```json
{
  "scope": "compacted",
  "stream": "logs",
  "dates": [
    "2026-06-01",
    "2026-06-02",
    "2026-06-03",
    "2026-06-10"
  ]
}
```

**With specific stream:**
```bash
openchami-logq-query inspect dates --stream events
```

### Inspect Schema

See the schema of your data:

```bash
openchami-logq-query inspect schema
```

Output:
```json
{
  "scope": "compacted",
  "stream": "logs",
  "schema": {
    "ts": "TIMESTAMP",
    "host": "VARCHAR",
    "level": "VARCHAR",
    "facility": "VARCHAR",
    "severity": "VARCHAR",
    "msg": "VARCHAR",
    "data": "JSON",
    "parse_error": "VARCHAR"
  }
}
```

### Inspect Configuration

See your current configuration:

```bash
openchami-logq-query inspect config
```

Output:
```json
{
  "s3_endpoint": "http://localhost:7070",
  "s3_region": "us-east-1",
  "s3_bucket_raw": "openchami-logs-raw",
  "s3_bucket_compacted": "openchami-logs-daily",
  "s3_ssl": false,
  "format": "json",
  "scope": "compacted",
  "stream": "logs"
}
```

---

## Output Formats

### JSON (Default)

Output as JSON array:

```bash
openchami-logq-query sql --format json \
  "SELECT * FROM SOURCES LIMIT 2"
```

Output:
```json
[
  {"ts":"2026-06-10T12:00:00Z","host":"node01","level":"INFO","msg":"Started"},
  {"ts":"2026-06-10T12:00:01Z","host":"node02","level":"ERROR","msg":"Failed"}
]
```

**Pros:**
- ✅ Valid JSON (can parse as array)
- ✅ Easy to pretty-print

**Cons:**
- ⚠️ Entire result in memory
- ⚠️ Large results can be slow

### NDJSON (Streaming)

Output as newline-delimited JSON:

```bash
openchami-logq-query sql --format ndjson \
  "SELECT * FROM SOURCES LIMIT 2"
```

Output:
```json
{"ts":"2026-06-10T12:00:00Z","host":"node01","level":"INFO","msg":"Started"}
{"ts":"2026-06-10T12:00:01Z","host":"node02","level":"ERROR","msg":"Failed"}
```

**Pros:**
- ✅ Streaming (constant memory)
- ✅ Fast for large results
- ✅ Easy to process line-by-line

**Cons:**
- ⚠️ Not valid JSON array
- ⚠️ Requires NDJSON-aware tools

### Output to File

Save output to file:

```bash
openchami-logq-query sql --output results.json \
  "SELECT * FROM SOURCES LIMIT 1000"
```

Or use shell redirection:

```bash
openchami-logq-query sql \
  "SELECT * FROM SOURCES LIMIT 1000" > results.json
```

### Pipe to jq

Process with jq:

```bash
# Pretty-print JSON
openchami-logq-query sql "SELECT * FROM SOURCES LIMIT 10" | jq '.'

# Filter with jq
openchami-logq-query sql "SELECT * FROM SOURCES LIMIT 100" | \
  jq '.[] | select(.level=="ERROR")'

# Extract field
openchami-logq-query sql "SELECT * FROM SOURCES LIMIT 100" | \
  jq -r '.[] | .msg'
```

### Pipe to Other Tools

**Count lines (NDJSON):**
```bash
openchami-logq-query sql --format ndjson \
  "SELECT * FROM SOURCES WHERE level='ERROR'" | wc -l
```

**Grep for pattern:**
```bash
openchami-logq-query sql --format ndjson \
  "SELECT * FROM SOURCES" | grep -i "timeout"
```

**Process with awk:**
```bash
openchami-logq-query sql --format ndjson \
  "SELECT * FROM SOURCES" | \
  awk -F'"' '{print $4}'  # Extract second field
```

---

## Advanced Usage

### Multi-Stream Queries

Query both logs and events:

```bash
openchami-logq-query sql --stream logs,events \
  "SELECT * FROM SOURCES LIMIT 10"
```

This creates a UNION of both streams.

### Raw Data Queries

Query raw NDJSON (before compaction):

```bash
openchami-logq-query sql --scope raw \
  "SELECT * FROM SOURCES WHERE ts > '2026-06-10T12:00:00Z'"
```

**Use case:** Query very recent data (not yet compacted).

### All Data Queries

Query both raw and compacted:

```bash
openchami-logq-query sql --scope all \
  "SELECT * FROM SOURCES"
```

**Use case:** Complete view of all data.

### Complex Aggregations

**Error rate over time:**
```bash
openchami-logq-query sql \
  "SELECT
     DATE_TRUNC('hour', ts) as hour,
     COUNT(*) as total,
     SUM(CASE WHEN level = 'ERROR' THEN 1 ELSE 0 END) as errors,
     ROUND(100.0 * errors / total, 2) as error_rate
   FROM SOURCES
   GROUP BY hour
   ORDER BY hour"
```

**Top error messages:**
```bash
openchami-logq-query sql \
  "SELECT
     msg,
     COUNT(*) as count,
     COUNT(DISTINCT host) as affected_hosts
   FROM SOURCES
   WHERE level = 'ERROR'
   GROUP BY msg
   ORDER BY count DESC
   LIMIT 10"
```

### Window Functions

**Rank hosts by log volume:**
```bash
openchami-logq-query sql \
  "SELECT
     host,
     COUNT(*) as log_count,
     RANK() OVER (ORDER BY COUNT(*) DESC) as rank
   FROM SOURCES
   GROUP BY host
   ORDER BY rank"
```

**Running total:**
```bash
openchami-logq-query sql \
  "SELECT
     DATE_TRUNC('day', ts) as day,
     COUNT(*) as daily_count,
     SUM(COUNT(*)) OVER (ORDER BY day) as cumulative_count
   FROM SOURCES
   GROUP BY day
   ORDER BY day"
```

### Performance Optimization

**Use WHERE to filter early:**
```bash
# Good: Filter before aggregation
openchami-logq-query sql \
  "SELECT host, COUNT(*)
   FROM SOURCES
   WHERE ts > '2026-06-10'
   GROUP BY host"

# Bad: Filter after aggregation
openchami-logq-query sql \
  "SELECT host, COUNT(*)
   FROM (SELECT * FROM SOURCES WHERE ts > '2026-06-10')
   GROUP BY host"
```

**Select only needed columns:**
```bash
# Good: Only select needed columns
openchami-logq-query sql \
  "SELECT ts, host, msg FROM SOURCES"

# Bad: Select all then filter
openchami-logq-query sql \
  "SELECT * FROM SOURCES" | jq '.[] | {ts, host, msg}'
```

**Use LIMIT for exploration:**
```bash
# Good: Limit results
openchami-logq-query sql \
  "SELECT * FROM SOURCES LIMIT 100"

# Bad: Return everything
openchami-logq-query sql \
  "SELECT * FROM SOURCES"
```

---

## Troubleshooting

### Error: S3 access denied

**Problem:**
```
Error: S3 access denied: Access Denied
```

**Solution:**
1. Check credentials:
   ```bash
   echo $S3_ACCESS_KEY
   echo $S3_SECRET_KEY
   ```

2. Verify S3 access:
   ```bash
   aws s3 ls s3://openchami-logs-daily \
     --endpoint-url=$S3_ENDPOINT
   ```

3. Check IAM policy (user needs `s3:GetObject`, `s3:ListBucket`)

### Error: Connection refused

**Problem:**
```
Error: Connection refused
```

**Solution:**
1. Check endpoint:
   ```bash
   echo $S3_ENDPOINT
   curl $S3_ENDPOINT  # Should return something
   ```

2. Check VersityGW is running:
   ```bash
   systemctl status versitygw
   ```

3. Check SSL setting:
   ```bash
   # If using HTTP (not HTTPS)
   export S3_SSL="false"
   ```

### Error: No data returned

**Problem:**
Query returns empty result.

**Solution:**
1. Check available dates:
   ```bash
   openchami-logq-query inspect dates
   ```

2. Verify data exists:
   ```bash
   aws s3 ls s3://openchami-logs-daily/logs/ \
     --endpoint-url=$S3_ENDPOINT
   ```

3. Check scope:
   ```bash
   # Try raw data
   openchami-logq-query sql --scope raw \
     "SELECT * FROM SOURCES LIMIT 10"
   ```

### Error: Invalid SQL

**Problem:**
```
Error: Parser Error: syntax error at or near "..."
```

**Solution:**
1. Check SQL syntax:
   ```bash
   # Test with simple query first
   openchami-logq-query sql "SELECT 1"
   ```

2. Quote strings properly:
   ```bash
   # Good
   "SELECT * FROM SOURCES WHERE level = 'ERROR'"

   # Bad
   "SELECT * FROM SOURCES WHERE level = ERROR"
   ```

3. Use SOURCES placeholder:
   ```bash
   # Good
   "SELECT * FROM SOURCES"

   # Bad
   "SELECT * FROM logs"  # Don't use table name directly
   ```

### Slow Queries

**Problem:**
Query takes a long time.

**Solution:**
1. Add WHERE clause to filter early:
   ```bash
   # Filter by date first
   "SELECT * FROM SOURCES WHERE ts > '2026-06-10'"
   ```

2. Use LIMIT for exploration:
   ```bash
   "SELECT * FROM SOURCES LIMIT 100"
   ```

3. Check data size:
   ```bash
   # How much data are you querying?
   openchami-logq-query sql \
     "SELECT COUNT(*) FROM SOURCES"
   ```

---

## Best Practices

### 1. Always Use WHERE for Time Ranges

**Good:**
```bash
openchami-logq-query sql \
  "SELECT * FROM SOURCES
   WHERE ts >= '2026-06-10' AND ts < '2026-06-11'"
```

**Why:** DuckDB can skip reading Parquet files outside the time range (partition pruning).

### 2. Use LIMIT for Exploration

**Good:**
```bash
# First, explore with LIMIT
openchami-logq-query sql \
  "SELECT * FROM SOURCES LIMIT 100"

# Then, run full query if needed
openchami-logq-query sql \
  "SELECT * FROM SOURCES WHERE level='ERROR'"
```

**Why:** Avoid accidentally downloading gigabytes of data.

### 3. Select Only Needed Columns

**Good:**
```bash
openchami-logq-query sql \
  "SELECT ts, host, msg FROM SOURCES"
```

**Bad:**
```bash
openchami-logq-query sql \
  "SELECT * FROM SOURCES"
```

**Why:** DuckDB only reads the columns you select (columnar format).

### 4. Use NDJSON for Large Results

**Good:**
```bash
openchami-logq-query sql --format ndjson \
  "SELECT * FROM SOURCES" > large-results.ndjson
```

**Why:** Streaming format, constant memory.

### 5. Use Reports for Common Queries

**Good:**
```bash
openchami-logq-query report run find-all-service-errors
```

**Why:** Consistent, tested, and documented.

### 6. Check Available Dates First

**Good:**
```bash
# Check dates first
openchami-logq-query inspect dates

# Then query specific date
openchami-logq-query sql \
  "SELECT * FROM SOURCES WHERE ts >= '2026-06-10'"
```

**Why:** Avoid querying dates with no data.

### 7. Use Environment Variables

**Good:**
```bash
# Set once
export S3_ENDPOINT="http://localhost:7070"
export S3_ACCESS_KEY="..."

# Use many times
openchami-logq-query sql "SELECT * FROM SOURCES"
```

**Why:** Don't repeat configuration on every command.

### 8. Test Queries on Compacted Data

**Good:**
```bash
# Test on compacted data first (faster)
openchami-logq-query sql --scope compacted \
  "SELECT * FROM SOURCES LIMIT 10"

# Then query raw if needed
openchami-logq-query sql --scope raw \
  "SELECT * FROM SOURCES WHERE ts > NOW() - INTERVAL 1 HOUR"
```

**Why:** Compacted data is faster to query (Parquet vs NDJSON).

---

## FAQ

### Q: How do I query the last hour of logs?

```bash
openchami-logq-query sql --scope raw \
  "SELECT * FROM SOURCES
   WHERE ts > NOW() - INTERVAL 1 HOUR"
```

Use `--scope raw` for recent data (not yet compacted).

### Q: How do I count errors by host?

```bash
openchami-logq-query sql \
  "SELECT host, COUNT(*) as error_count
   FROM SOURCES
   WHERE level = 'ERROR'
   GROUP BY host
   ORDER BY error_count DESC"
```

### Q: How do I search for a specific message?

```bash
openchami-logq-query sql \
  "SELECT * FROM SOURCES
   WHERE msg LIKE '%timeout%'
   LIMIT 100"
```

### Q: How do I query both logs and events?

```bash
openchami-logq-query sql --stream logs,events \
  "SELECT * FROM SOURCES LIMIT 10"
```

### Q: How do I save results to a file?

```bash
openchami-logq-query sql --output results.json \
  "SELECT * FROM SOURCES WHERE level='ERROR'"
```

Or use shell redirection:
```bash
openchami-logq-query sql \
  "SELECT * FROM SOURCES WHERE level='ERROR'" > results.json
```

### Q: How do I query a specific date?

```bash
openchami-logq-query sql \
  "SELECT * FROM SOURCES
   WHERE ts >= '2026-06-10'
   AND ts < '2026-06-11'"
```

### Q: How do I find unique values?

```bash
# Unique hosts
openchami-logq-query sql \
  "SELECT DISTINCT host FROM SOURCES"

# Unique levels
openchami-logq-query sql \
  "SELECT DISTINCT level FROM SOURCES"
```

### Q: How do I join logs with external data?

You can't directly join with external data, but you can:

1. Export logs to file
2. Load into DuckDB
3. Join with external data

```bash
# Export logs
openchami-logq-query sql --format ndjson \
  "SELECT * FROM SOURCES" > logs.ndjson

# Use DuckDB CLI
duckdb <<SQL
CREATE TABLE logs AS
  SELECT * FROM read_json('logs.ndjson');
CREATE TABLE external AS
  SELECT * FROM read_csv('external.csv');
SELECT * FROM logs JOIN external ON logs.host = external.hostname;
SQL
```

### Q: How do I monitor query performance?

Add timing:

```bash
time openchami-logq-query sql "SELECT COUNT(*) FROM SOURCES"
```

Output:
```
{"count":1000000}

real    0m2.341s
user    0m0.123s
sys     0m0.045s
```

### Q: How do I query CloudEvents?

```bash
openchami-logq-query sql --stream events \
  "SELECT type, source, COUNT(*)
   FROM SOURCES
   GROUP BY type, source"
```

### Q: How do I extract JSON fields?

```bash
openchami-logq-query sql \
  "SELECT
     json_extract_string(data, '$.user') as user,
     COUNT(*) as count
   FROM SOURCES
   WHERE data IS NOT NULL
   GROUP BY user"
```

### Q: Can I use DuckDB functions?

Yes! All DuckDB SQL functions are available:

```bash
# Date functions
openchami-logq-query sql \
  "SELECT DATE_TRUNC('hour', ts) as hour, COUNT(*)
   FROM SOURCES GROUP BY hour"

# String functions
openchami-logq-query sql \
  "SELECT UPPER(host), LOWER(level) FROM SOURCES"

# Math functions
openchami-logq-query sql \
  "SELECT ROUND(AVG(LENGTH(msg)), 2) FROM SOURCES"
```

See [DuckDB Functions](https://duckdb.org/docs/sql/functions/overview) for complete list.

---

## Next Steps

- **[Architecture Guide](ARCHITECTURE.md)** - Understand system design
- **[Developer Guide](DEVELOPMENT.md)** - Contribute to the project
- **[Operations Guide](OPERATIONS.md)** - Deploy and operate
- **[API Reference](API_REFERENCE.md)** - Complete CLI reference

---

**Need help?** Open an issue on [GitHub](https://github.com/OpenCHAMI/legendary-funicular/issues)!
