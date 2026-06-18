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
SELECT * FROM SOURCES WHERE ts::TIMESTAMP > '2026-06-10'

-- Query engine converts to:
SELECT * FROM read_parquet([
  's3://openchami-logs-daily/logs/*.parquet',
  's3://openchami-logs-raw/logs/*.parquet',
  ...
])
WHERE ts::TIMESTAMP > '2026-06-10'
```

### Scope Control

Control which data SOURCES points to:

```bash
# Query compacted Parquet files (fast)
openchami-logq-query sql --scope compacted "SELECT * FROM SOURCES"

# Query raw NDJSON files (fast)
openchami-logq-query sql --scope raw "SELECT * FROM SOURCES"

# Query both raw and compacted (default, slower, all data)
openchami-logq-query sql --scope all "SELECT * FROM SOURCES"
```

**When to use each:**
- `compacted`: To query historical data (fast, data older than 1 day)
- `raw`: To query recent data not yet compacted (fast, last 24 hours)
- `all`: To query all data (default, slower)

### Stream Selection

Query different log types:

```bash
# Query syslog data (default)
openchami-logq-query sql --stream logs "SELECT * FROM SOURCES"

# Query CloudEvents data
openchami-logq-query sql --stream events "SELECT * FROM SOURCES"
```

---

## Schema and Fields

### Inspect Your Schema

See what fields are available:

```bash
$ openchami-logq-query inspect schema --stream logs
{"column_name":"ts","column_type":"VARCHAR","is_nullable":"YES"}
{"column_name":"service","column_type":"VARCHAR","is_nullable":"YES"}
{"column_name":"host","column_type":"VARCHAR","is_nullable":"YES"}
{"column_name":"level","column_type":"VARCHAR","is_nullable":"YES"}
{"column_name":"msg","column_type":"VARCHAR","is_nullable":"YES"}
{"column_name":"data","column_type":"VARCHAR","is_nullable":"YES"}
{"column_name":"xname","column_type":"VARCHAR","is_nullable":"YES"}
{"column_name":"component_id","column_type":"VARCHAR","is_nullable":"YES"}
{"column_name":"node_id","column_type":"VARCHAR","is_nullable":"YES"}
{"column_name":"trace_id","column_type":"VARCHAR","is_nullable":"YES"}
{"column_name":"request_id","column_type":"VARCHAR","is_nullable":"YES"}
{"column_name":"request_uri","column_type":"VARCHAR","is_nullable":"YES"}
{"column_name":"request_user","column_type":"VARCHAR","is_nullable":"YES"}
{"column_name":"parse_error","column_type":"VARCHAR","is_nullable":"YES"}
{"column_name":"payload_json","column_type":"VARCHAR","is_nullable":"YES"}
{"column_name":"date","column_type":"DATE","is_nullable":"YES"}

$ openchami-logq-query inspect schema --stream events
{"column_name":"ts","column_type":"VARCHAR","is_nullable":"YES"}
{"column_name":"transport_method","column_type":"VARCHAR","is_nullable":"YES"}
{"column_name":"transport_metadata","column_type":"VARCHAR","is_nullable":"YES"}
{"column_name":"cloudevent","column_type":"VARCHAR","is_nullable":"YES"}
{"column_name":"cloudevent_id","column_type":"VARCHAR","is_nullable":"YES"}
{"column_name":"cloudevent_source","column_type":"VARCHAR","is_nullable":"YES"}
{"column_name":"cloudevent_type","column_type":"VARCHAR","is_nullable":"YES"}
{"column_name":"cloudevent_specversion","column_type":"VARCHAR","is_nullable":"YES"}
{"column_name":"cloudevent_binding","column_type":"VARCHAR","is_nullable":"YES"}
{"column_name":"xname","column_type":"VARCHAR","is_nullable":"YES"}
{"column_name":"component_id","column_type":"VARCHAR","is_nullable":"YES"}
{"column_name":"node_id","column_type":"VARCHAR","is_nullable":"YES"}
{"column_name":"trace_id","column_type":"VARCHAR","is_nullable":"YES"}
{"column_name":"request_id","column_type":"VARCHAR","is_nullable":"YES"}
{"column_name":"request_uri","column_type":"VARCHAR","is_nullable":"YES"}
{"column_name":"request_user","column_type":"VARCHAR","is_nullable":"YES"}
{"column_name":"raw","column_type":"VARCHAR","is_nullable":"YES"}
{"column_name":"parse_error","column_type":"VARCHAR","is_nullable":"YES"}
{"column_name":"payload_json","column_type":"VARCHAR","is_nullable":"YES"}
{"column_name":"date","column_type":"DATE","is_nullable":"YES"}
```

### Handling Missing Fields

Fields may be NULL if not present in the log:

```sql
-- Filter out NULLs
SELECT * FROM SOURCES WHERE msg IS NOT NULL

-- Provide defaults
SELECT COALESCE(level, 'UNKNOWN') as level FROM SOURCES
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

## Next Steps

- **[Architecture Guide](ARCHITECTURE.md)** - Understand system design
- **[Developer Guide](DEVELOPMENT.md)** - Contribute to the project

---

**Need help?** Open an issue on [GitHub](https://github.com/OpenCHAMI/legendary-funicular/issues)!
