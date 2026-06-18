<!--
SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC

SPDX-License-Identifier: MIT
-->

# Architecture

Technical architecture and design decisions for openchami-logq.

## Table of Contents

1. [System Overview](#system-overview)
2. [Architecture Principles](#architecture-principles)
3. [Component Architecture](#component-architecture)
4. [Data Flow](#data-flow)
5. [Storage Architecture](#storage-architecture)
6. [Query Architecture](#query-architecture)
7. [Technology Choices](#technology-choices)

---

## System Overview

### Purpose

openchami-logq is a lightweight log lake system for HPC environments providing:
- **Cheap storage** for all logs and events
- **Fast queries** using SQL
- **Schema flexibility** for evolving log formats
- **Zero maintenance** - no clusters, no indices

### Design Philosophy

1. **Storage First** - Never lose data, optimize for cost
2. **Query Second** - Fast enough for ad-hoc analysis
3. **Simple Always** - Minimal setup and maintenance
4. **HPC Aware** - Built for distributed systems

### System Diagram

```mermaid
graph TD
    A[Log Sources] -->|syslog/events| B[Collector<br/>Vector]
    B -->|NDJSON| C[S3 Raw Bucket<br/>openchami-logs-raw]
    C -->|daily| D[Compactor<br/>openchami-logq-compactor]
    D -->|Parquet| E[S3 Compacted Bucket<br/>openchami-logs-daily]
    E -->|SQL queries| F[Query CLI<br/>openchami-logq-query]
    F -->|JSON/NDJSON| G[User]
```

---

## Architecture Principles

### 1. Append-Only Storage

**Decision:** Never delete or modify raw logs.

**Rationale:**
- Logs are immutable evidence
- Storage is cheap
- Deletion risks losing critical data

**Implementation:**
- Raw logs stored as-is in NDJSON
- Compaction creates new files, doesn't modify originals
- Original deletion only after successful compaction

### 2. Schema-on-Read

**Decision:** No schema validation during ingestion.

**Rationale:**
- Log formats evolve over time
- Schema validation can cause data loss
- Better to store everything and parse later

**Implementation:**
- Collector writes raw JSON/syslog without validation
- Compactor does best-effort field extraction
- Query engine handles missing/extra fields gracefully

### 3. Separation of Concerns

**Decision:** Separate components for collection, storage, compaction, and querying.

**Rationale:**
- Each component can scale independently
- Failures in one don't affect others
- Easier to maintain and debug

**Implementation:**
- Collector: Vector (external dependency)
- Storage: S3-compatible (external dependency)
- Compactor: Go binary, runs on schedule
- Query: Go CLI, runs on-demand

### 4. Direct S3 Access

**Decision:** Query Parquet files directly from S3 without copying.

**Rationale:**
- No local storage required
- Leverages DuckDB's S3 integration
- Reduces operational complexity

**Implementation:**
- DuckDB reads Parquet from S3 via HTTP range requests
- Only fetches required row groups and columns
- No intermediate database or cache

---

## Component Architecture

### Collector (Vector)

**Responsibility:** Accept logs and write to S3 raw bucket.

**Key Characteristics:**
- External dependency (not part of this project)
- Configured via YAML
- Writes NDJSON format
- Batches writes (5-10MB or 5 minutes)

**Configuration:**

Check out the [default configuration](collector/vector.d/)

### Compactor

**Responsibility:** Convert raw NDJSON to optimized Parquet.

**Key Characteristics:**
- Go binary
- Runs daily (default configuration uses a Systemd timer)
- Stateless (no local database)
- Streaming architecture (constant memory)

**Architecture:**

```mermaid
graph LR
    A[S3 Raw] -->|stream| B[Reader]
    B --> C[Parser<br/>syslog/cloudevent]
    C --> D[Parquet Writer]
    D -->|upload| E[S3 Compacted]
    E -->|success| F[Delete Raw]

    style A fill:#770077
    style E fill:#007777
```

**Parsing Strategy:**
- Select format based on the S3 bucket prefix (logs vs. events)
- Extract common fields: `ts`, `host`, `level`, `msg`
- Store unknown fields in `data` JSON column
- Never fail on parse errors (store as-is)

### Query Engine

**Responsibility:** Execute SQL queries against Parquet and compressed raw NDJSON files.

**Key Characteristics:**
- Go binary with embedded DuckDB
- Stateless (no local database)
- Reads directly from S3
- Outputs JSON or NDJSON

**Architecture:**

```mermaid
graph LR
    A[User SQL] --> B[openchami-logq-query]
    B --> C[DuckDB]
    C -->|HTTP range requests| D[S3 Parquet]
    D --> C
    C --> B
    B -->|JSON/NDJSON| E[stdout]

    style D fill:#e1ffe1
```

**Query Pattern:**
```sql
-- SOURCES is replaced with actual S3 paths
SELECT * FROM SOURCES WHERE ts::TIMESTAMP > '2026-06-10'

-- Becomes:
SELECT * FROM read_parquet('s3://<bucket-name>/logs/**/*.parquet')
```

---

## Data Flow

### Ingestion Flow

```mermaid
sequenceDiagram
    participant App as Application
    participant Collector as Vector
    participant S3 as S3 Raw Bucket

    App->>Collector: Send log (syslog/event)
    Collector->>Collector: Buffer (5min or 10MB)
    Collector->>S3: Write NDJSON
    Note over S3: logs/../1781821498-ab1730b7-cfcb-491b-8b30-acff12603e3c.ndjson.zst
```

**Path Pattern:** `logs/date=YYYY-MM-DD/hour=HH/<UUID>.ndjson.zst`

**File Size:** Typically 5-10MB (5 minutes of logs)

### Compaction Flow

```mermaid
sequenceDiagram
    participant Cron as Cron/Timer
    participant Compactor as Compactor
    participant Raw as S3 Raw
    participant Compacted as S3 Compacted

    Cron->>Compactor: Run daily at 2 AM
    Compactor->>Raw: List logs/**/*.ndjson.zst
    Raw-->>Compactor: File list
    loop For each file
        Raw->>Compactor: Stream NDJSON
        Compactor->>Compactor: Parse & convert to Parquet
    end
    Compactor->>Compacted: Upload Parquet
    Compactor->>Raw: Delete processed files
```

**Path Pattern:** `logs/date=YYYY-MM-DD/<UUID>.parquet`

**Compression:** Typically 10:1 (100MB NDJSON → 10MB Parquet)

### Query Flow

```mermaid
sequenceDiagram
    participant User
    participant CLI as openchami-logq-query
    participant DuckDB
    participant S3 as S3 Compacted + Raw

    User->>CLI: SQL query
    CLI->>CLI: Replace SOURCES with S3 paths
    CLI->>DuckDB: Execute query
    DuckDB->>S3: HTTP range request (row groups)
    S3-->>DuckDB: Parquet data
    DuckDB-->>CLI: Result rows
    CLI->>User: JSON/NDJSON output
```

**Optimization:** DuckDB only fetches required columns and row groups (partition pruning).

---

## Storage Architecture

### S3 Bucket Structure

```
openchami-logs-raw/
├── events
│   └── date=2026-06-18
│       └── hour=22
│           └── 1781820712-6e7df24a-114e-46c0-8fc8-d01070eae668.ndjson.zst
└── logs
    └── date=2026-06-18
        └── hour=22
            └── 1781820714-f2966c24-8a82-42da-9c82-c8c995caf504.ndjson.zst

openchami-logs-daily/
├── events
│   └── date=2026-03-23
│   │   └── fa34270d-1dc3-4981-921e-ad17b4b64dd1.parquet
└── logs
    └── date=2026-04-06
        └── 4046851d-6872-490e-9ded-7a4ca459f4fc.parquet
```

### Query Data Format

**Syslog:**
```json
{
  "component_id": null,
  "data": {...},
  "date": "2026-06-10",
  "host": "srv00",
  "level": "info",
  "msg": "<message content>",
  "node_id": null,
  "parse_error": null,
  "payload_json": "{...}\n",
  "request_id": null,
  "request_uri": null,
  "request_user": null,
  "service": "versitygw",
  "trace_id": null,
  "ts": "2026-03-20T14:12:40Z",
  "xname": ""
}
```

**CloudEvents:**
```json
{
  "cloudevent": {
    "data": {
      "hello": "world",
      "mode": "structured"
    },
    "datacontenttype": "application/json",
    "id": "3a5e17ff-dadd-4efb-8c60-6f7bf83a7db7",
    "source": "example/uri",
    "specversion": "1.0",
    "time": "2026-03-23T18:04:55.161504723Z",
    "type": "org.openchami.cetester.dummy"
  },
  "cloudevent_binding": "structured",
  "cloudevent_id": "3a5e17ff-dadd-4efb-8c60-6f7bf83a7db7",
  "cloudevent_source": "example/uri",
  "cloudevent_specversion": "1.0",
  "cloudevent_type": "org.openchami.cetester.dummy",
  "component_id": null,
  "date": "2026-03-23",
  "node_id": null,
  "parse_error": null,
  "payload_json": "{...}\n",
  "raw": "{...}",
  "request_id": null,
  "request_uri": null,
  "request_user": null,
  "trace_id": null,
  "transport_metadata": "{\"accept_encoding\":\"gzip\",\"content_length\":\"248\",\"content_type\":\"application/cloudevents+json\",\"host\":\"localhost:8910\",\"path\":\"/\",\"user_agent\":\"Go-http-client/1.1\"}",
  "transport_method": "http",
  "ts": "2026-06-06T18:04:55.162900372Z",
  "xname": null
}
```

### IAM Model

Three IAM users with minimal permissions:

**log-writer** (Collector):
- `s3:PutObject` on raw bucket
- `s3:ListBucket` on raw bucket

**log-compactor** (Compactor):
- `s3:GetObject`, `s3:ListBucket` on raw bucket
- `s3:PutObject` on compacted bucket
- `s3:DeleteObject` on raw bucket (post-compaction)

**log-reader** (Query):
- `s3:GetObject`, `s3:ListBucket` on both raw and compacted buckets

---

## Technology Choices

### Why S3?

**Pros:**
- Industry standard (AWS, MinIO, VersityGW all compatible)
- Cheap storage
- No operational overhead (managed service or simple self-hosted)
- HTTP-based (works through firewalls)

**Cons:**
- Higher latency than local disk
- Request costs (mitigated by batching)

**Decision:** Storage cost and simplicity outweigh latency concerns for log analytics.

### Why NDJSON?

**Pros:**
- Human-readable (debugging friendly)
- Streamable (process line-by-line)
- Schema-flexible (JSON supports any structure)
- Universal (every language has JSON support)

**Cons:**
- Larger than binary formats
- Slower to parse than binary

**Decision:** Schema flexibility and debuggability more important than size/speed for raw logs.

### Why Parquet?

**Pros:**
- Columnar format (efficient for analytics)
- Excellent compression (10:1 typical)
- Self-describing (schema embedded)
- Industry standard (Spark, DuckDB, Snowflake all support)
- Partition pruning (skip files/row groups)

**Cons:**
- Not human-readable
- Write-once (can't append)

**Decision:** Perfect fit for compacted, immutable log data.

### Why DuckDB?

**Pros:**
- Embedded (no server to manage)
- Native S3 support
- Full SQL support
- Columnar engine (fast aggregations)
- Small binary
- MIT license

**Cons:**
- Not distributed (single-node only)
- Embedded (can't share connections)

**Decision:** Simplicity and SQL compatibility more important than distributed queries for log analytics.

### Why Go?

**Pros:**
- Single binary (no dependencies)
- Fast compilation
- Good stdlib (HTTP, JSON, CLI)
- Cross-platform
- Memory safe

**Cons:**
- Verbose error handling
- No generics (until 1.18+)

**Decision:** Operational simplicity (single binary) is critical for HPC deployments.

### Why Vector?

**Decision:** Use existing, battle-tested collectors rather than building our own.

**Rationale:**
- Mature projects with wide adoption
- Support many log sources (syslog, journald, files, etc.)
- Built-in buffering and retry logic
- Active development and community

---

## Related Documentation

- **[User Guide](USER_GUIDE.md)** - Query patterns and SQL examples
- **[Developer Guide](DEVELOPMENT.md)** - Development setup

---

**Questions?** Open an issue on [GitHub](https://github.com/OpenCHAMI/legendary-funicular/issues)!
