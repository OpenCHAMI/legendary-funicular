<!--
SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC

SPDX-License-Identifier: MIT
-->

# User Guide

Advanced usage guide for openchami-logq. For basic usage and installation, see the [main README](../README.md).

## Table of Contents

1. [Advanced SQL Examples](#advanced-sql-examples)
2. [DuckDB Best Practices](#duckdb-best-practices)
3. [FAQ](#faq)

---

## Advanced SQL Examples

### Window Functions

Calculate running totals:
```bash
openchami-logq-query sql "
  SELECT
    ts,
    host,
    COUNT(*) OVER (PARTITION BY host ORDER BY ts) as running_count
  FROM SOURCES
  WHERE ts >= '2026-06-10'
  ORDER BY ts"
```

Find the last error for each host:
```bash
openchami-logq-query sql "
  SELECT * FROM (
    SELECT
      *,
      ROW_NUMBER() OVER (PARTITION BY host ORDER BY ts DESC) as rn
    FROM SOURCES
    WHERE level = 'ERROR'
  ) WHERE rn = 1"
```

### Common Table Expressions (CTEs)

Multi-step analysis:
```bash
openchami-logq-query sql "
  WITH error_counts AS (
    SELECT host, COUNT(*) as errors
    FROM SOURCES
    WHERE level = 'ERROR'
    GROUP BY host
  ),
  total_counts AS (
    SELECT host, COUNT(*) as total
    FROM SOURCES
    GROUP BY host
  )
  SELECT
    e.host,
    e.errors,
    t.total,
    ROUND(100.0 * e.errors / t.total, 2) as error_rate
  FROM error_counts e
  JOIN total_counts t ON e.host = t.host
  ORDER BY error_rate DESC"
```

### JSON Field Extraction

Extract nested fields:
```bash
openchami-logq-query sql "
  SELECT
    json_extract_string(data, '$.user.name') as username,
    json_extract_string(data, '$.user.id') as user_id,
    json_extract(data, '$.metadata.tags') as tags,
    COUNT(*) as count
  FROM SOURCES
  WHERE data IS NOT NULL
  GROUP BY username, user_id, tags"
```

### Array Operations

Work with array fields:
```bash
openchami-logq-query sql "
  SELECT
    host,
    UNNEST(json_extract(data, '$.errors')) as error
  FROM SOURCES
  WHERE json_extract(data, '$.errors') IS NOT NULL"
```

### Date/Time Operations

Group by custom intervals:
```bash
openchami-logq-query sql "
  SELECT
    DATE_TRUNC('minute', ts) as minute,
    DATE_TRUNC('hour', ts) as hour,
    DATE_TRUNC('day', ts) as day,
    COUNT(*) as count
  FROM SOURCES
  WHERE ts >= NOW() - INTERVAL 7 DAY
  GROUP BY minute, hour, day"
```

Calculate time differences:
```bash
openchami-logq-query sql "
  SELECT
    host,
    ts,
    LAG(ts) OVER (PARTITION BY host ORDER BY ts) as prev_ts,
    EPOCH(ts - LAG(ts) OVER (PARTITION BY host ORDER BY ts)) as seconds_since_last
  FROM SOURCES
  WHERE level = 'ERROR'
  ORDER BY host, ts"
```

### String Operations

Pattern matching and extraction:
```bash
openchami-logq-query sql "
  SELECT
    msg,
    REGEXP_EXTRACT(msg, 'error code: ([0-9]+)', 1) as error_code,
    REGEXP_MATCHES(msg, 'timeout|failed|error') as is_error
  FROM SOURCES
  WHERE msg IS NOT NULL
  LIMIT 100"
```

### Aggregations

Statistical functions:
```bash
openchami-logq-query sql "
  SELECT
    host,
    COUNT(*) as count,
    MIN(ts) as first_seen,
    MAX(ts) as last_seen,
    APPROX_COUNT_DISTINCT(msg) as unique_messages,
    PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY LENGTH(msg)) as median_msg_length
  FROM SOURCES
  GROUP BY host"
```

### Cross-Stream Queries

Join logs and events:
```bash
openchami-logq-query sql --stream logs,events "
  SELECT
    CASE
      WHEN type IS NOT NULL THEN 'event'
      ELSE 'log'
    END as source_type,
    COALESCE(type, level) as category,
    COUNT(*) as count
  FROM SOURCES
  GROUP BY source_type, category
  ORDER BY count DESC"
```

---

## DuckDB Best Practices

### 1. Always Use WHERE for Time Ranges

**Good:**
```bash
openchami-logq-query sql "
  SELECT * FROM SOURCES
  WHERE ts >= '2026-06-10' AND ts < '2026-06-11'"
```

**Why:** DuckDB can skip reading Parquet files outside the time range (partition pruning). This dramatically improves query performance.

**Impact:** Can reduce query time from minutes to seconds for large datasets.

### 2. Use LIMIT for Exploration

**Good:**
```bash
# First, explore with LIMIT
openchami-logq-query sql "SELECT * FROM SOURCES LIMIT 100"

# Then, run full query if needed
openchami-logq-query sql "SELECT * FROM SOURCES WHERE level='ERROR'"
```

**Why:** Avoid accidentally downloading gigabytes of data while exploring.

### 3. Select Only Needed Columns

**Good:**
```bash
openchami-logq-query sql "SELECT ts, host, msg FROM SOURCES"
```

**Bad:**
```bash
openchami-logq-query sql "SELECT * FROM SOURCES"
```

**Why:** DuckDB only reads the columns you select (columnar format). Selecting fewer columns = less data transferred from S3 = faster queries.

**Impact:** Can reduce data transfer by 90%+ for wide tables.

### 4. Use NDJSON for Large Results

**Good:**
```bash
openchami-logq-query sql --format ndjson "
  SELECT * FROM SOURCES" > large-results.ndjson
```

**Why:**
- Streaming format with constant memory usage
- Can process results line-by-line with `jq` or other tools
- No need to load entire result set into memory

**Impact:** Can query datasets larger than available RAM.

### 5. Use Built-in Reports

**Good:**
```bash
openchami-logq-query report run find-all-service-errors
```

**Why:** Reports are:
- Pre-tested and optimized
- Documented with expected output
- Consistent across users

### 6. Check Available Dates First

**Good:**
```bash
# Check dates first
openchami-logq-query inspect dates

# Then query specific date
openchami-logq-query sql "
  SELECT * FROM SOURCES WHERE ts >= '2026-06-10'"
```

**Why:** Avoid querying dates with no data or querying more data than necessary.

### 7. Use Environment Variables

**Good:**
```bash
# Set once
export S3_ENDPOINT="http://localhost:7070"
export S3_ACCESS_KEY="..."
export S3_SECRET_KEY="..."

# Use many times
openchami-logq-query sql "SELECT * FROM SOURCES"
```

**Why:** Don't repeat configuration on every command.

### 8. Test Queries on Compacted Data First

**Good:**
```bash
# Test on compacted data first (faster)
openchami-logq-query sql --scope compacted "
  SELECT * FROM SOURCES LIMIT 10"

# Then query raw if needed
openchami-logq-query sql --scope raw "
  SELECT * FROM SOURCES WHERE ts > NOW() - INTERVAL 1 HOUR"
```

**Why:**
- Compacted data (Parquet) is 10-20x faster to query than raw data (NDJSON)
- Use raw data only for very recent logs (last few hours)

### 9. Use APPROX Functions for Large Datasets

**Good:**
```bash
openchami-logq-query sql "
  SELECT APPROX_COUNT_DISTINCT(host) as approx_hosts FROM SOURCES"
```

**Better than:**
```bash
openchami-logq-query sql "
  SELECT COUNT(DISTINCT host) as exact_hosts FROM SOURCES"
```

**Why:** APPROX functions are much faster and use less memory. Error rate is typically <2%.

### 10. Optimize GROUP BY Order

**Good:**
```bash
openchami-logq-query sql "
  SELECT host, level, COUNT(*)
  FROM SOURCES
  GROUP BY host, level"  # host has higher cardinality
```

**Less optimal:**
```bash
openchami-logq-query sql "
  SELECT level, host, COUNT(*)
  FROM SOURCES
  GROUP BY level, host"  # level has lower cardinality
```

**Why:** Grouping by high-cardinality columns first can improve query performance.

---

## FAQ

### Q: How do I query the last hour of logs?

```bash
openchami-logq-query sql --scope raw "
  SELECT * FROM SOURCES
  WHERE ts > NOW() - INTERVAL 1 HOUR"
```

Use `--scope raw` for recent data that hasn't been compacted yet.

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
openchami-logq-query sql "
  SELECT * FROM SOURCES
  WHERE msg LIKE '%timeout%'
  LIMIT 100"
```

For case-insensitive search:
```bash
openchami-logq-query sql "
  SELECT * FROM SOURCES
  WHERE LOWER(msg) LIKE '%timeout%'
  LIMIT 100"
```

### Q: How do I query both logs and events?

```bash
openchami-logq-query sql --stream logs,events "
  SELECT * FROM SOURCES LIMIT 10"
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

# Unique levels
openchami-logq-query sql "SELECT DISTINCT level FROM SOURCES"

# Count of unique values
openchami-logq-query sql "SELECT COUNT(DISTINCT host) FROM SOURCES"
```

### Q: How do I join logs with external data?

DuckDB can read external files directly in queries:

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

### Q: How do I monitor query performance?

Add timing with the `time` command:

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

### Q: Can I use all DuckDB functions?

Yes! All DuckDB SQL functions are available:

```bash
# Date functions
openchami-logq-query sql "
  SELECT DATE_TRUNC('hour', ts) as hour, COUNT(*)
  FROM SOURCES GROUP BY hour"

# String functions
openchami-logq-query sql "
  SELECT UPPER(host), LOWER(level) FROM SOURCES"

# Math functions
openchami-logq-query sql "
  SELECT ROUND(AVG(LENGTH(msg)), 2) FROM SOURCES"

# JSON functions
openchami-logq-query sql "
  SELECT json_extract_string(data, '$.key') FROM SOURCES"
```

See [DuckDB Functions](https://duckdb.org/docs/sql/functions/overview) for complete reference.

### Q: How do I handle NULL values?

```bash
# Filter out NULLs
openchami-logq-query sql "
  SELECT * FROM SOURCES WHERE msg IS NOT NULL"

# Replace NULLs with default
openchami-logq-query sql "
  SELECT COALESCE(msg, 'no message') as message FROM SOURCES"

# Count NULLs
openchami-logq-query sql "
  SELECT
    COUNT(*) as total,
    COUNT(msg) as non_null_msg,
    COUNT(*) - COUNT(msg) as null_msg
  FROM SOURCES"
```

### Q: How do I debug slow queries?

1. **Check how much data you're scanning:**
```bash
openchami-logq-query sql "
  SELECT COUNT(*), MIN(ts), MAX(ts) FROM SOURCES"
```

2. **Add WHERE clause to limit time range:**
```bash
# Bad: scans all data
openchami-logq-query sql "SELECT * FROM SOURCES WHERE level='ERROR'"

# Good: scans only one day
openchami-logq-query sql "
  SELECT * FROM SOURCES
  WHERE ts >= '2026-06-10' AND level='ERROR'"
```

3. **Select fewer columns:**
```bash
# Bad: reads all columns
openchami-logq-query sql "SELECT * FROM SOURCES"

# Good: reads only needed columns
openchami-logq-query sql "SELECT ts, host, msg FROM SOURCES"
```

4. **Use EXPLAIN to see query plan:**
```bash
openchami-logq-query sql "
  EXPLAIN SELECT * FROM SOURCES WHERE ts >= '2026-06-10'"
```

### Q: How do I query very large result sets?

For result sets larger than available RAM:

1. **Use NDJSON format** (streaming):
```bash
openchami-logq-query sql --format ndjson "
  SELECT * FROM SOURCES" | jq -c 'select(.level=="ERROR")'
```

2. **Process in chunks** with LIMIT/OFFSET:
```bash
# Process 10,000 rows at a time
for offset in 0 10000 20000 30000; do
  openchami-logq-query sql "
    SELECT * FROM SOURCES
    LIMIT 10000 OFFSET $offset" >> results.ndjson
done
```

3. **Use aggregate queries** to reduce data:
```bash
# Instead of downloading all rows, aggregate first
openchami-logq-query sql "
  SELECT host, level, COUNT(*), MIN(ts), MAX(ts)
  FROM SOURCES
  GROUP BY host, level"
```

---

## Next Steps

- **[Main README](../README.md)** - Installation and basic usage
- **[Architecture Guide](ARCHITECTURE.md)** - System design and technical details
- **[Operations Guide](OPERATIONS.md)** - Production deployment and operations
- **[Developer Guide](DEVELOPMENT.md)** - Contributing and development setup

---

**Need help?** Open an issue on [GitHub](https://github.com/OpenCHAMI/legendary-funicular/issues)!
