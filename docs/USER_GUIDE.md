<!--
SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC

SPDX-License-Identifier: MIT
-->

# User Guide

Advanced usage guide for openchami-logq. For installation and basic usage, see the [main README](../README.md).

## Table of Contents

1. [Understanding SOURCES](#understanding-sources)
2. [Schema and Fields](#schema-and-fields)
3. [JSON Field Extraction](#json-field-extraction)
4. [DuckDB Best Practices](#duckdb-best-practices)
5. [FAQ](#faq)

---

## Understanding SOURCES

### What is SOURCES?

`SOURCES` is a placeholder that gets replaced with actual S3 paths to your log files.

**Example:**
```sql
-- You write:
SELECT * FROM SOURCES WHERE ts > '2026-06-10'

-- Query engine converts to:
SELECT * FROM read_parquet([
  's3://openchami-logs-daily/logs/2026-06-10.parquet',
  's3://openchami-logs-daily/logs/2026-06-11.parquet',
  ...
])
WHERE ts > '2026-06-10'
```

### Scope Control

Control which data SOURCES points to:

```bash
# Query compacted Parquet files (fast, default)
openchami-logq-query sql --scope compacted "SELECT * FROM SOURCES"

# Query raw NDJSON files (recent data, slower)
openchami-logq-query sql --scope raw "SELECT * FROM SOURCES"

# Query both raw and compacted (complete data)
openchami-logq-query sql --scope all "SELECT * FROM SOURCES"
```

**When to use each:**
- `compacted`: Default for most queries (fast, data older than 1 day)
- `raw`: Recent data not yet compacted (last 24 hours)
- `all`: Complete historical analysis (slower)

### Stream Selection

Query different log types:

```bash
# Query syslog data (default)
openchami-logq-query sql --stream logs "SELECT * FROM SOURCES"

# Query CloudEvents data
openchami-logq-query sql --stream events "SELECT * FROM SOURCES"

# Query both
openchami-logq-query sql --stream logs,events "SELECT * FROM SOURCES"
```

---

## Schema and Fields

### Inspect Your Schema

See what fields are available:

```bash
openchami-logq-query inspect schema --stream logs
```

**Example output:**
```
┌────────────┬─────────┬─────────┬─────────┬─────────┬─────────┐
│ column_name│ column_type│ null│ key│ default│ extra│
├────────────┼─────────┼─────────┼─────────┼─────────┼─────────┤
│ ts         │ TIMESTAMP│ YES  │     │         │         │
│ host       │ VARCHAR  │ YES  │     │         │         │
│ level      │ VARCHAR  │ YES  │     │         │         │
│ msg        │ VARCHAR  │ YES  │     │         │         │
│ data       │ JSON     │ YES  │     │         │         │
└────────────┴─────────┴─────────┴─────────┴─────────┴─────────┘
```

### Common Fields

**Syslog logs:**
- `ts` (TIMESTAMP): Log timestamp
- `host` (VARCHAR): Hostname
- `level` (VARCHAR): Log level (INFO, ERROR, WARN, DEBUG)
- `msg` (VARCHAR): Log message
- `data` (JSON): Additional fields

**CloudEvents:**
- `ts` (TIMESTAMP): Event time
- `source` (VARCHAR): Event source
- `type` (VARCHAR): Event type
- `specversion` (VARCHAR): CloudEvents version
- `data` (JSON): Event payload

### Handling Missing Fields

Fields may be NULL if not present in the log:

```sql
-- Filter out NULLs
SELECT * FROM SOURCES WHERE msg IS NOT NULL

-- Provide defaults
SELECT COALESCE(level, 'UNKNOWN') as level FROM SOURCES

-- Count NULLs
SELECT COUNT(*) - COUNT(msg) as null_messages FROM SOURCES
```

---

## JSON Field Extraction

Logs often contain nested JSON in the `data` field. DuckDB provides JSON functions to extract values.

### Basic Extraction

```sql
-- Extract string value
SELECT json_extract_string(data, '$.user') as user FROM SOURCES

-- Extract numeric value
SELECT json_extract(data, '$.count') as count FROM SOURCES

-- Extract nested object
SELECT json_extract(data, '$.metadata.tags') as tags FROM SOURCES
```

### Example: User Activity

Given logs with this structure:
```json
{"ts": "2026-06-10T12:00:00Z", "data": {"user": "alice", "action": "login", "ip": "192.168.1.1"}}
```

Query:
```bash
openchami-logq-query sql "
  SELECT
    ts,
    json_extract_string(data, '$.user') as user,
    json_extract_string(data, '$.action') as action,
    json_extract_string(data, '$.ip') as ip
  FROM SOURCES
  WHERE json_extract_string(data, '$.action') = 'login'
  ORDER BY ts DESC
  LIMIT 10"
```

### Array Operations

Extract and expand arrays:

```sql
-- Unnest array into rows
SELECT
  host,
  UNNEST(json_extract(data, '$.errors')) as error
FROM SOURCES
WHERE json_extract(data, '$.errors') IS NOT NULL
```

### JSON Path Syntax

DuckDB uses JSONPath syntax:

- `$.field` - Top-level field
- `$.nested.field` - Nested field
- `$.array[0]` - Array element
- `$.array[*]` - All array elements

**Full reference:** [DuckDB JSON Functions](https://duckdb.org/docs/sql/functions/json)

---

## DuckDB Best Practices

### 1. Always Filter by Timestamp

**Why:** Enables partition pruning - DuckDB skips entire Parquet files outside your time range.

```sql
-- Good: Filters files before reading
SELECT * FROM SOURCES
WHERE ts >= '2026-06-10' AND ts < '2026-06-11'

-- Bad: Reads all files then filters
SELECT * FROM SOURCES LIMIT 100
```

**Impact:** Can reduce query time from minutes to seconds.

### 2. Select Only Needed Columns

**Why:** Columnar format means only selected columns are read from S3.

```sql
-- Good: Reads 2 columns
SELECT ts, msg FROM SOURCES

-- Bad: Reads all columns
SELECT * FROM SOURCES
```

**Impact:** Can reduce data transfer by 80%+.

### 3. Use LIMIT for Exploration

**Why:** Prevents accidentally downloading gigabytes of data.

```sql
-- Good: Explore first
SELECT * FROM SOURCES LIMIT 100

-- Then: Run full query if needed
SELECT * FROM SOURCES WHERE level='ERROR'
```

### 4. Use NDJSON for Large Results

**Why:** Streaming format with constant memory usage.

```bash
# Good: Stream large results
openchami-logq-query sql --format ndjson "
  SELECT * FROM SOURCES" > large-results.ndjson

# Process with jq
cat large-results.ndjson | jq 'select(.level=="ERROR")'
```

### 5. Check Available Dates First

**Why:** Avoid querying dates with no data.

```bash
# Check dates first
openchami-logq-query inspect dates --stream logs

# Then query specific date
openchami-logq-query sql "
  SELECT * FROM SOURCES WHERE ts >= '2026-06-10'"
```

### 6. Use APPROX Functions for Large Datasets

**Why:** Much faster, error rate typically <2%.

```sql
-- Good: Fast approximation
SELECT APPROX_COUNT_DISTINCT(host) as approx_hosts FROM SOURCES

-- Slower: Exact count
SELECT COUNT(DISTINCT host) as exact_hosts FROM SOURCES
```

---

## FAQ

### Q: What SQL functions can I use?

**A:** All DuckDB SQL functions! This includes:
- Aggregations: `COUNT`, `SUM`, `AVG`, `MIN`, `MAX`, `STDDEV`, etc.
- Window functions: `ROW_NUMBER`, `RANK`, `LAG`, `LEAD`, etc.
- String functions: `UPPER`, `LOWER`, `SUBSTRING`, `REGEXP_MATCHES`, etc.
- Date functions: `DATE_TRUNC`, `DATE_ADD`, `DATE_DIFF`, etc.
- JSON functions: `json_extract`, `json_extract_string`, etc.
- Math functions: `ROUND`, `CEIL`, `FLOOR`, `ABS`, etc.

**Full reference:** [DuckDB Functions](https://duckdb.org/docs/sql/functions/overview)

### Q: Can I use window functions and CTEs?

**A:** Yes! DuckDB supports full SQL including:
- Window functions (`OVER` clauses)
- Common Table Expressions (`WITH` clauses)
- Subqueries
- Joins (including self-joins)
- `UNION`, `INTERSECT`, `EXCEPT`

**Examples:** [DuckDB SQL Introduction](https://duckdb.org/docs/sql/introduction)

### Q: How do I query the last hour of logs?

```bash
openchami-logq-query sql --scope raw "
  SELECT * FROM SOURCES
  WHERE ts > NOW() - INTERVAL 1 HOUR"
```

Use `--scope raw` for recent data not yet compacted.

### Q: How do I count errors by host?

```bash
openchami-logq-query sql "
  SELECT host, COUNT(*) as error_count
  FROM SOURCES
  WHERE level = 'ERROR'
  GROUP BY host
  ORDER BY error_count DESC"
```

### Q: How do I search for a specific message?

```bash
# Case-sensitive
openchami-logq-query sql "
  SELECT * FROM SOURCES
  WHERE msg LIKE '%timeout%'
  LIMIT 100"

# Case-insensitive
openchami-logq-query sql "
  SELECT * FROM SOURCES
  WHERE LOWER(msg) LIKE '%timeout%'
  LIMIT 100"
```

### Q: How do I save results to a file?

```bash
# Using --output flag
openchami-logq-query sql --output results.json "
  SELECT * FROM SOURCES WHERE level='ERROR'"

# Using shell redirection
openchami-logq-query sql "
  SELECT * FROM SOURCES WHERE level='ERROR'" > results.json

# For large results, use NDJSON
openchami-logq-query sql --format ndjson "
  SELECT * FROM SOURCES WHERE level='ERROR'" > results.ndjson
```

### Q: How do I query a specific date range?

```bash
openchami-logq-query sql "
  SELECT * FROM SOURCES
  WHERE ts >= '2026-06-10' AND ts < '2026-06-11'"
```

### Q: How do I find unique values?

```bash
# Unique hosts
openchami-logq-query sql "SELECT DISTINCT host FROM SOURCES"

# Count unique values
openchami-logq-query sql "SELECT COUNT(DISTINCT host) FROM SOURCES"

# Unique values with counts
openchami-logq-query sql "
  SELECT host, COUNT(*) as count
  FROM SOURCES
  GROUP BY host
  ORDER BY count DESC"
```

### Q: How do I join logs with external data?

DuckDB can read external CSV/JSON/Parquet files directly:

```bash
openchami-logq-query sql "
  SELECT
    s.*,
    h.location,
    h.rack
  FROM SOURCES s
  JOIN read_csv('hosts.csv') h ON s.host = h.hostname
  WHERE s.level = 'ERROR'"
```

**Supported formats:** CSV, JSON, NDJSON, Parquet
**Reference:** [DuckDB Data Import](https://duckdb.org/docs/data/overview)

### Q: How do I monitor query performance?

```bash
# Use time command
time openchami-logq-query sql "SELECT COUNT(*) FROM SOURCES"

# Output:
# {"count":1000000}
# real    0m2.341s
```

For slow queries:
1. Check data volume: `SELECT COUNT(*) FROM SOURCES`
2. Add timestamp filter: `WHERE ts >= '2026-06-10'`
3. Select fewer columns: `SELECT ts, msg` instead of `SELECT *`
4. Use `EXPLAIN`: `EXPLAIN SELECT * FROM SOURCES`

### Q: How do I handle very large result sets?

For results larger than available RAM:

```bash
# 1. Use NDJSON format (streaming)
openchami-logq-query sql --format ndjson "
  SELECT * FROM SOURCES" | jq -c 'select(.level=="ERROR")'

# 2. Use aggregation to reduce data
openchami-logq-query sql "
  SELECT host, level, COUNT(*), MIN(ts), MAX(ts)
  FROM SOURCES
  GROUP BY host, level"

# 3. Process in batches with LIMIT/OFFSET
for offset in 0 10000 20000; do
  openchami-logq-query sql "
    SELECT * FROM SOURCES LIMIT 10000 OFFSET $offset"
done
```

### Q: What if I get "S3 access denied"?

```bash
# 1. Check credentials
aws s3 ls s3://openchami-logs-daily --endpoint-url=$S3_ENDPOINT

# 2. Verify configuration
openchami-logq-query inspect config

# 3. Check environment variables
echo $S3_ENDPOINT
echo $S3_ACCESS_KEY
```

### Q: Can I use this with BI tools?

Yes! Export to formats that BI tools understand:

```bash
# Export to CSV
openchami-logq-query sql "
  SELECT * FROM SOURCES WHERE level='ERROR'" \
  | jq -r '(.[0] | keys_unsorted) as $keys | $keys, map([.[ $keys[] ]])[] | @csv' \
  > errors.csv

# Export to Parquet (via DuckDB CLI)
duckdb -c "
  COPY (SELECT * FROM SOURCES WHERE level='ERROR')
  TO 'errors.parquet' (FORMAT PARQUET)"
```

---

## Next Steps

- **[Architecture Guide](ARCHITECTURE.md)** - Understand system design
- **[Operations Guide](OPERATIONS.md)** - Deploy and operate
- **[Developer Guide](DEVELOPMENT.md)** - Contribute to the project

---

**Need help?** Open an issue on [GitHub](https://github.com/OpenCHAMI/legendary-funicular/issues)!
