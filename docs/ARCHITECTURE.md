<!--
SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC

SPDX-License-Identifier: MIT
-->

# Architecture Documentation

**Project:** openchami-logq
**Version:** 1.0
**Last Updated:** June 10, 2026

---

## Table of Contents

1. [System Overview](#system-overview)
2. [Architecture Principles](#architecture-principles)
3. [Component Architecture](#component-architecture)
4. [Data Flow](#data-flow)
5. [Storage Architecture](#storage-architecture)
6. [Query Architecture](#query-architecture)
7. [Compaction Architecture](#compaction-architecture)
8. [Security Architecture](#security-architecture)
9. [Scalability & Performance](#scalability--performance)
10. [Technology Choices](#technology-choices)

---

## System Overview

### Purpose

openchami-logq is a lightweight log lake system designed for HPC environments. It provides:

- **Cheap storage** for all logs and events
- **Fast queries** using SQL
- **Schema flexibility** to handle evolving log formats
- **Zero maintenance** - no clusters, no indices

### Design Philosophy

1. **Storage First** - Never lose data, optimize cost
2. **Query Second** - Fast enough for ad-hoc analysis
3. **Simple Always** - No complex setup or maintenance
4. **HPC Aware** - Built for xnames, services, and distributed systems

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Data Sources                              │
│  ┌─────────┐  ┌────────────┐  ┌──────────┐  ┌──────────────┐  │
│  │ Syslog  │  │ CloudEvents│  │  Vector  │  │  FluentBit   │  │
│  └────┬────┘  └─────┬──────┘  └────┬─────┘  └──────┬───────┘  │
└───────┼─────────────┼──────────────┼────────────────┼──────────┘
        │             │              │                │
        └─────────────┴──────────────┴────────────────┘
                              │
                              v
                  ┌───────────────────────┐
                  │   Collector Layer     │
                  │  (Vector/FluentBit)   │
                  │  - Accepts all logs   │
                  │  - No validation      │
                  │  - NDJSON output      │
                  └───────────┬───────────┘
                              │
                              v
        ┌─────────────────────────────────────────┐
        │         Storage Layer (S3)               │
        │  ┌────────────────┐  ┌────────────────┐│
        │  │  Raw Bucket    │  │Compacted Bucket││
        │  │   (NDJSON)     │  │   (Parquet)    ││
        │  │  - Minutely    │  │   - Daily      ││
        │  │  - Compressed  │  │   - Optimized  ││
        │  └───────┬────────┘  └────────┬───────┘│
        └──────────┼────────────────────┼────────┘
                   │                    │
                   v                    │
          ┌────────────────┐            │
          │   Compactor    │            │
          │  - Reads raw   │────────────┘
          │  - Parses logs │
          │  - Writes      │
          │    Parquet     │
          └────────────────┘
                                        │
                                        v
                              ┌──────────────────┐
                              │   Query Engine   │
                              │    (DuckDB)      │
                              │  - SQL queries   │
                              │  - Direct S3     │
                              │  - No indexing   │
                              └─────────┬────────┘
                                        │
                                        v
                              ┌──────────────────┐
                              │   CLI Interface  │
                              │ - openchami-logq │
                              │ - JSON output    │
                              │ - Reports        │
                              └──────────────────┘
```

---

## Architecture Principles

### 1. Never Lose Data

**Principle:** Logs are written to storage before any processing.

**Implementation:**
- Collectors write directly to S3 (no intermediate queues)
- NDJSON format (one log per line, never corrupts entire file)
- No schema validation (accept everything)
- Compaction only deletes after successful Parquet write

**Trade-offs:**
- ✅ Zero data loss
- ✅ Simple recovery
- ⚠️ Higher storage costs initially
- ⚠️ Raw data requires parsing at query time

### 2. Optimize for Cost

**Principle:** Logs are cheap to store, expensive to query.

**Implementation:**
- S3 storage (~$0.023/GB/month)
- Compress NDJSON with zstd (3x reduction)
- Convert to Parquet daily (10x reduction)
- Move old data to S3 Glacier (90% cost reduction)

**Economics:**
```
1TB logs/month:
- Raw NDJSON:     1000 GB × $0.023 = $23/month
- Compressed:      333 GB × $0.023 = $7.66/month
- Parquet:         100 GB × $0.004 = $0.40/month (S3 IA)
- 1 year old:       10 GB × $0.001 = $0.01/month (Glacier)

Total: ~$8/month vs $1000+/month for traditional systems
```

### 3. Schema Flexibility

**Principle:** Log formats evolve, systems shouldn't break.

**Implementation:**
- Store raw JSON (preserves all fields)
- Extract known fields best-effort
- Unknown fields preserved in `data` column
- No schema migrations required

**Example:**
```json
// Old format
{"ts": "...", "host": "...", "msg": "..."}

// New format (works without changes)
{"ts": "...", "host": "...", "msg": "...", "trace_id": "...", "xname": "..."}
```

### 4. Query Performance "Good Enough"

**Principle:** Don't optimize for speed, optimize for simplicity.

**Implementation:**
- DuckDB (embedded, no cluster)
- Columnar Parquet (only read needed columns)
- Direct S3 access (no data movement)
- Predicate pushdown (filter at storage layer)

**Performance:**
- Simple queries: <100ms
- Aggregations: <1s
- Full scans: <5s (1GB data)

**Not optimized for:**
- ❌ Real-time dashboards (use Prometheus)
- ❌ High-frequency queries (use time-series DB)
- ❌ Sub-second latency (use streaming)

### 5. Zero Maintenance

**Principle:** No clusters, no indices, no tuning.

**Implementation:**
- Embedded DuckDB (no server)
- S3 storage (managed by provider)
- Stateless compactor (cron job)
- No data migration

**Operations:**
- Deploy: Copy binary + set env vars
- Monitor: Check compactor logs
- Backup: S3 replication (built-in)
- Scale: Add more S3 buckets

---

## Component Architecture

### 1. Collector (Vector/FluentBit)

**Responsibility:** Ingest logs and write to S3.

**Design:**
```
┌──────────────┐
│   Syslog     │
│   Input      │──┐
└──────────────┘  │
                  ├──> ┌──────────────┐     ┌──────────────┐
┌──────────────┐  │    │   Parser     │     │   S3 Sink    │
│ CloudEvents  │──┼───>│ (optional)   │────>│   (NDJSON)   │
│   Input      │  │    └──────────────┘     └──────────────┘
└──────────────┘  │
                  │
┌──────────────┐  │
│   HTTP       │──┘
│   Input      │
└──────────────┘
```

**Configuration:**
- **Inputs:** syslog (TCP/UDP), CloudEvents (HTTP), JSON (HTTP)
- **Transform:** Add timestamp, normalize format
- **Output:** S3 bucket, NDJSON format, minutely rotation
- **Buffering:** 10MB or 60 seconds (whichever first)

**Key Files:**
- `collector/vector.yaml` - Vector configuration
- `collector/fluent-bit.conf` - FluentBit configuration

### 2. Compactor

**Responsibility:** Convert NDJSON → Parquet daily.

**Design:**
```
┌─────────────────────────────────────────────────────────┐
│                    Compactor Process                     │
│                                                          │
│  1. Discovery                                           │
│     ┌────────────────────────────────────┐             │
│     │ List S3 objects for date           │             │
│     │ Filter by prefix (logs/ or events/)│             │
│     └────────────┬───────────────────────┘             │
│                  │                                      │
│  2. Pipeline Setup                                     │
│     ┌────────────v───────────────────────┐             │
│     │ Create streaming pipeline          │             │
│     │ - Reader goroutine (S3 → parse)    │             │
│     │ - Writer goroutine (Parquet → S3)  │             │
│     │ - Error channels                   │             │
│     └────────────┬───────────────────────┘             │
│                  │                                      │
│  3. Reader (Producer)                                  │
│     ┌────────────v───────────────────────┐             │
│     │ For each S3 object:                │             │
│     │   - Stream download                │             │
│     │   - Decompress (zstd)              │             │
│     │   - Parse lines (syslog/cloudevent)│             │
│     │   - Convert to Parquet schema      │             │
│     │   - Write to pipe                  │             │
│     └────────────┬───────────────────────┘             │
│                  │                                      │
│  4. Writer (Consumer)                                  │
│     ┌────────────v───────────────────────┐             │
│     │ Read from pipe                     │             │
│     │ Upload Parquet to S3               │             │
│     │ Log bytes written                  │             │
│     └────────────┬───────────────────────┘             │
│                  │                                      │
│  5. Cleanup                                            │
│     ┌────────────v───────────────────────┐             │
│     │ If success:                        │             │
│     │   - Delete source NDJSON files     │             │
│     │ If failure:                        │             │
│     │   - Keep source files for retry    │             │
│     └────────────────────────────────────┘             │
└─────────────────────────────────────────────────────────┘
```

**Key Features:**
- **Streaming:** Constant memory (no full load)
- **Concurrent:** Reader and writer run in parallel
- **Safe:** Only deletes after successful write
- **Resumable:** Failed compactions can retry

**Key Files:**
- `compactor/main.go` - Entry point
- `compactor/compaction.go` - Main compaction logic
- `compactor/internal/record/` - Parsers (syslog, cloudevent)
- `compactor/internal/pipeline/` - Streaming pipeline
- `compactor/internal/zio/` - Compression (zstd)

### 3. Query Engine

**Responsibility:** Execute SQL queries on Parquet files.

**Design:**
```
┌─────────────────────────────────────────────────────────┐
│                    Query Process                         │
│                                                          │
│  1. Configuration                                       │
│     ┌────────────────────────────────────┐             │
│     │ Load S3 credentials                │             │
│     │ Parse CLI flags                    │             │
│     │ Build source paths                 │             │
│     └────────────┬───────────────────────┘             │
│                  │                                      │
│  2. SQL Engine Setup                                   │
│     ┌────────────v───────────────────────┐             │
│     │ Create DuckDB connection           │             │
│     │ Configure S3 secret                │             │
│     │ Prepare query:                     │             │
│     │   - Replace SOURCES placeholder    │             │
│     │   - Add UNION for multiple sources │             │
│     │   - Wrap with to_json if needed    │             │
│     └────────────┬───────────────────────┘             │
│                  │                                      │
│  3. Query Execution                                    │
│     ┌────────────v───────────────────────┐             │
│     │ Execute prepared query             │             │
│     │ DuckDB:                            │             │
│     │   - Reads Parquet from S3          │             │
│     │   - Applies predicates             │             │
│     │   - Returns result stream          │             │
│     └────────────┬───────────────────────┘             │
│                  │                                      │
│  4. Result Processing                                  │
│     ┌────────────v───────────────────────┐             │
│     │ For each row:                      │             │
│     │   - Scan (JSON or struct)          │             │
│     │   - Encode (JSON or NDJSON)        │             │
│     │   - Write to output                │             │
│     └────────────┬───────────────────────┘             │
│                  │                                      │
│  5. Cleanup                                            │
│     ┌────────────v───────────────────────┐             │
│     │ Close result set                   │             │
│     │ Close encoder                      │             │
│     │ Close DuckDB connection            │             │
│     └────────────────────────────────────┘             │
└─────────────────────────────────────────────────────────┘
```

**Key Features:**
- **Direct S3 Access:** No data movement
- **Columnar Reading:** Only read needed columns
- **Predicate Pushdown:** Filter at storage layer
- **Streaming Output:** Constant memory

**Key Files:**
- `query/main.go` - Entry point
- `query/cmd/sql/` - SQL command
- `query/cmd/report/` - Report system
- `query/internal/sql/` - DuckDB integration
- `query/internal/render/` - Output formatting

---

## Data Flow

### Ingestion Flow

```
1. Log Generated
   ├─> Application writes log
   └─> Sent to collector (syslog/HTTP)

2. Collector Receives
   ├─> Parse (if syslog)
   ├─> Add timestamp
   ├─> Convert to NDJSON
   └─> Buffer (10MB or 60s)

3. Write to S3 (Raw Bucket)
   ├─> Compress with zstd
   ├─> Key: logs/YYYY-MM-DD/HH-MM-SS-uuid.ndjson.zst
   └─> ACL: log-writer (write-only)

4. Raw Storage
   ├─> Retention: 7 days
   ├─> Size: ~1GB/day (typical)
   └─> Cost: ~$0.023/GB/month
```

### Compaction Flow

```
1. Compactor Starts (Daily 2 AM)
   ├─> Date: yesterday
   └─> Prefix: logs/2026-06-09/

2. List Raw Files
   ├─> S3 ListObjects
   ├─> Filter by prefix
   └─> Result: [file1.ndjson.zst, file2.ndjson.zst, ...]

3. Stream Processing
   ├─> For each file:
   │   ├─> Download (streaming)
   │   ├─> Decompress (zstd)
   │   ├─> Parse (line-by-line)
   │   ├─> Extract fields
   │   └─> Write to Parquet (streaming)
   └─> Memory: constant (no full load)

4. Write Parquet
   ├─> Compress with snappy
   ├─> Key: logs/date=2026-06-09/uuid.parquet
   ├─> ACL: log-compactor (write), log-reader (read)
   └─> Size: ~100MB (10x compression)

5. Cleanup
   ├─> Verify Parquet upload
   ├─> Delete raw NDJSON files
   └─> Log completion
```

### Query Flow

```
1. User Executes Query
   └─> openchami-logq-query sql "SELECT ..."

2. Parse Configuration
   ├─> Load S3 credentials
   ├─> Parse --scope, --stream, --format
   └─> Build source paths

3. Prepare Query
   ├─> Replace SOURCES placeholder
   ├─> Add UNION for multiple streams
   ├─> Wrap with to_json() if needed
   └─> Example: SELECT * FROM read_parquet('s3://bucket/logs/*.parquet')

4. Execute with DuckDB
   ├─> DuckDB connects to S3
   ├─> Reads Parquet metadata
   ├─> Applies WHERE filters
   ├─> Reads only needed columns
   ├─> Returns result stream
   └─> Performance: ~1s for 1GB

5. Format Output
   ├─> Scan each row
   ├─> Encode (JSON or NDJSON)
   ├─> Stream to stdout
   └─> Memory: constant (no buffering)
```

---

## Storage Architecture

### S3 Bucket Structure

```
openchami-logs-raw/
├── logs/
│   ├── 2026-06-09/
│   │   ├── 00-15-30-abc123.ndjson.zst
│   │   ├── 00-16-30-def456.ndjson.zst
│   │   └── ... (96 files/day @ 15min intervals)
│   └── 2026-06-10/
│       └── ...
└── events/
    ├── 2026-06-09/
    │   └── ...
    └── 2026-06-10/
        └── ...

openchami-logs-daily/
├── logs/
│   ├── date=2026-06-09/
│   │   └── abc123-def456.parquet
│   └── date=2026-06-10/
│       └── ...
└── events/
    ├── date=2026-06-09/
    │   └── ...
    └── date=2026-06-10/
        └── ...
```

### File Formats

**Raw NDJSON:**
```json
{"ts":"2026-06-10T12:00:00Z","host":"node01","level":"INFO","msg":"Started"}
{"ts":"2026-06-10T12:00:01Z","host":"node02","level":"ERROR","msg":"Failed"}
```

**Compressed NDJSON (zstd):**
- Compression ratio: ~3:1
- Streaming: yes
- Seekable: no

**Parquet Schema (Syslog):**
```
ts: TIMESTAMP
host: STRING
level: STRING
facility: STRING
severity: STRING
msg: STRING
data: JSON  // Original full record
parse_error: STRING  // If parsing failed
```

**Parquet Schema (CloudEvents):**
```
ts: TIMESTAMP
type: STRING
source: STRING
specversion: STRING
id: STRING
cloudevent: JSON  // Original full record
parse_error: STRING  // If parsing failed
```

### Storage Lifecycle

```
Day 0-7:   Raw NDJSON (S3 Standard)
Day 7-30:  Parquet (S3 Standard)
Day 30-90: Parquet (S3 Infrequent Access)
Day 90+:   Parquet (S3 Glacier)
```

### IAM Policies

**log-writer (Collector):**
```json
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": ["s3:PutObject"],
    "Resource": "arn:aws:s3:::openchami-logs-raw/*"
  }]
}
```

**log-compactor:**
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": ["s3:GetObject", "s3:ListBucket", "s3:DeleteObject"],
      "Resource": "arn:aws:s3:::openchami-logs-raw/*"
    },
    {
      "Effect": "Allow",
      "Action": ["s3:PutObject"],
      "Resource": "arn:aws:s3:::openchami-logs-daily/*"
    }
  ]
}
```

**log-reader (Query):**
```json
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": ["s3:GetObject", "s3:ListBucket"],
    "Resource": "arn:aws:s3:::openchami-logs-daily/*"
  }]
}
```

---

## Query Architecture

### DuckDB Integration

**Why DuckDB?**
- Embedded (no server)
- Columnar (fast analytics)
- S3-native (direct access)
- SQL (familiar interface)

**Configuration:**
```sql
CREATE SECRET local_s3 (
  TYPE s3,
  PROVIDER config,
  KEY_ID 'access-key',
  SECRET 'secret-key',
  REGION 'us-east-1',
  ENDPOINT 'http://localhost:7070',
  USE_SSL false
);
```

### Query Patterns

**Pattern 1: Simple SELECT**
```sql
-- User query
SELECT * FROM SOURCES WHERE level = 'ERROR'

-- Transformed query
SELECT * exclude(data), json(data) as data
FROM read_parquet('s3://bucket/logs/date=2026-06-10/*.parquet')
WHERE level = 'ERROR'
```

**Pattern 2: Multi-Source UNION**
```sql
-- User query
SELECT * FROM SOURCES  -- with --stream logs,events

-- Transformed query
SELECT * FROM (
  SELECT * FROM read_parquet('s3://bucket/logs/*.parquet')
  UNION ALL BY NAME
  SELECT * FROM read_parquet('s3://bucket/events/*.parquet')
)
```

**Pattern 3: JSON Output**
```sql
-- User query with --format json
SELECT * FROM SOURCES LIMIT 10

-- Transformed query
SELECT to_json(t) FROM (
  SELECT * FROM read_parquet('s3://bucket/logs/*.parquet')
  LIMIT 10
) t
```

### Report System

**Architecture:**
```
┌──────────────────────────────────────────┐
│           Report Interface               │
│  - Name() string                         │
│  - Description() string                  │
│  - ParamSpecs() []ParamSpec              │
│  - BuildQueryString(params) (string, error)│
└──────────────────┬───────────────────────┘
                   │
                   │ implements
                   │
     ┌─────────────┴─────────────┐
     │                           │
┌────v────────────┐   ┌──────────v──────────┐
│ FindParseErrors │   │ FindServiceErrors   │
│  - No params    │   │  - No params        │
│  - Returns logs │   │  - Returns errors   │
│    with errors  │   │    by service       │
└─────────────────┘   └─────────────────────┘
```

**Registry Pattern:**
```go
// Register reports
func New() []report.Report {
    return []report.Report{
        &ReportFindParseErrors{},
        &ReportFindServiceErrors{},
    }
}

// Lookup by name
func Get(name string) (report.Report, error) {
    registry := New()
    // Find by name...
}
```

---

## Compaction Architecture

### Pipeline Design

**Streaming Pipeline:**
```
┌─────────┐   ┌──────────┐   ┌─────────┐   ┌──────────┐
│ S3 Get  │──>│ Decomp   │──>│ Parse   │──>│ Parquet  │
│         │   │ (zstd)   │   │ (line)  │   │ Writer   │
└─────────┘   └──────────┘   └─────────┘   └────┬─────┘
                                                  │
                                                  v
                                            ┌──────────┐
                                            │ S3 Put   │
                                            └──────────┘
```

**Goroutine Architecture:**
```
Main Goroutine
├─> Reader Goroutine (Producer)
│   ├─> For each S3 object
│   ├─> Download + decompress + parse
│   ├─> Write to pipe
│   └─> Send keys to channel
│
└─> Writer Goroutine (Consumer)
    ├─> Read from pipe
    ├─> Upload to S3
    └─> Send completion to channel

Wait for both goroutines
Check error channels
Delete source files if success
```

### Error Handling

**Philosophy:** Fail safe, keep source data.

**Implementation:**
```go
// Reader error
if err := transform(prev); err != nil {
    pipe.Close()           // Stop writer
    return nil, err        // Propagate error
}
// Source files NOT deleted

// Writer error
if err := sink.Put(key, pipe); err != nil {
    return nil, err        // Propagate error
}
// Source files NOT deleted

// Success
if err == nil {
    for _, key := range sourceKeys {
        source.Delete(key)  // Only delete on success
    }
}
```

### Parser Architecture

**Interface:**
```go
type Parser[T Record] interface {
    Parse(line []byte) (T, error)
}
```

**Implementations:**
- `SyslogParser` - RFC3164/RFC5424 syslog
- `CloudEventParser` - CloudEvents v1.0 JSON

**Fuzzing:**
- Both parsers have fuzz tests
- 1000+ generated inputs tested
- No crashes found

---

## Security Architecture

### Principle: Least Privilege

**Three Separate IAM Users:**
1. `log-writer` - Write-only to raw bucket
2. `log-compactor` - Read raw, write compacted, delete raw
3. `log-reader` - Read-only compacted bucket

### S3 Encryption

**At Rest:**
- SSE-S3 (default)
- Or SSE-KMS (customer managed keys)

**In Transit:**
- TLS 1.2+ (when S3_SSL=true)
- Or unencrypted (for internal VersityGW)

### Secrets Management

**Environment Variables:**
```bash
# Never commit these!
S3_ACCESS_KEY="..."
S3_SECRET_KEY="..."
```

**Production:**
- Use AWS Secrets Manager
- Or HashiCorp Vault
- Or Kubernetes Secrets

### Query Injection

**DuckDB Parameterization:**
```go
// SAFE: DuckDB handles S3 paths safely
engine.Query("SELECT * FROM read_parquet(?)", []string{s3Path})

// User SQL: Still risk of SQL injection
// TODO: Add query validation/sanitization
```

---

## Scalability & Performance

### Horizontal Scaling

**Collector:**
- Multiple instances (stateless)
- Load balancer in front
- Each writes to S3 independently

**Compactor:**
- One instance per date/prefix
- Can run multiple dates in parallel
- Idempotent (safe to retry)

**Query:**
- Stateless (embedded DuckDB)
- Unlimited concurrent queries
- Each query independent

### Performance Characteristics

**Write Throughput:**
- Collector: ~100MB/s per instance
- Bottleneck: S3 upload bandwidth

**Compaction Throughput:**
- ~10MB/s (limited by CPU for parsing)
- ~1 hour for 1TB of logs

**Query Latency:**
- Simple queries: <100ms
- Aggregations: <1s
- Full scans: <5s per 1GB

### Optimization Techniques

**1. Predicate Pushdown:**
```sql
-- DuckDB only reads matching rows from Parquet
SELECT * FROM logs WHERE ts > '2026-06-10' AND level = 'ERROR'
```

**2. Column Pruning:**
```sql
-- DuckDB only reads 'host' and 'msg' columns
SELECT host, msg FROM logs
```

**3. Partition Pruning:**
```
-- DuckDB only scans date=2026-06-10 partition
logs/date=2026-06-10/*.parquet
```

**4. Compression:**
- zstd for NDJSON (3:1)
- snappy for Parquet (10:1)
- Result: 30:1 overall

---

## Technology Choices

### Why S3?

**Pros:**
- ✅ Cheap storage
- ✅ Infinite scale
- ✅ Managed service
- ✅ Built-in replication
- ✅ Lifecycle policies

**Cons:**
- ⚠️ Eventual consistency
- ⚠️ Higher latency than local disk
- ⚠️ Per-request costs

**Alternatives Considered:**
- Local filesystem: Not scalable
- HDFS: Too complex
- Ceph: Requires cluster

### Why DuckDB?

**Pros:**
- ✅ Embedded (no server)
- ✅ Columnar (fast analytics)
- ✅ S3-native (direct access)
- ✅ SQL (familiar)
- ✅ Active development

**Cons:**
- ⚠️ Not distributed
- ⚠️ No real-time queries
- ⚠️ No indexing

**Alternatives Considered:**
- ClickHouse: Too complex
- Presto/Trino: Requires cluster
- Athena: AWS-only, expensive

### Why Parquet?

**Pros:**
- ✅ Columnar (fast queries)
- ✅ Compressed (10:1)
- ✅ Self-describing schema
- ✅ Industry standard

**Cons:**
- ⚠️ Not human-readable
- ⚠️ Requires tools to inspect

**Alternatives Considered:**
- CSV: Not typed, not compressed
- Avro: Row-based, slower queries
- ORC: Less tooling support

### Why Go?

**Pros:**
- ✅ Fast compilation
- ✅ Static binary (easy deploy)
- ✅ Great concurrency
- ✅ Good S3 libraries

**Cons:**
- ⚠️ Verbose error handling
- ⚠️ No generics (until 1.18)

**Alternatives Considered:**
- Python: Slower, requires runtime
- Rust: Steeper learning curve
- Java: Heavier runtime

---

## Deployment Patterns

### Pattern 1: Single Node

```
┌─────────────────────────────────┐
│        Single Server            │
│  ┌──────────┐  ┌──────────┐   │
│  │Collector │  │Compactor │   │
│  │(Vector)  │  │(cron)    │   │
│  └──────────┘  └──────────┘   │
│                                 │
│  Users SSH in and run:          │
│  $ openchami-logq-query sql ... │
└─────────────────────────────────┘
         │
         v
    ┌─────────┐
    │ S3/VGW  │
    └─────────┘
```

**Pros:** Simple, cheap
**Cons:** Single point of failure

### Pattern 2: Kubernetes

```
┌────────────────────────────────────┐
│         Kubernetes Cluster         │
│  ┌────────────┐  ┌──────────────┐ │
│  │ Collector  │  │  Compactor   │ │
│  │ DaemonSet  │  │  CronJob     │ │
│  └────────────┘  └──────────────┘ │
│                                    │
│  ┌────────────────────────────┐   │
│  │  Query Job (on-demand)     │   │
│  │  kubectl run query --image=│   │
│  └────────────────────────────┘   │
└────────────────────────────────────┘
         │
         v
    ┌─────────┐
    │   S3    │
    └─────────┘
```

**Pros:** Scalable, managed
**Cons:** Complex, expensive

### Pattern 3: Serverless (Future)

```
┌──────────────────────────────────┐
│      AWS Lambda/Functions        │
│  ┌────────────┐  ┌────────────┐ │
│  │ Collector  │  │ Compactor  │ │
│  │ (trigger)  │  │ (schedule) │ │
│  └────────────┘  └────────────┘ │
│                                  │
│  ┌────────────────────────────┐ │
│  │  Query API (HTTP endpoint) │ │
│  └────────────────────────────┘ │
└──────────────────────────────────┘
         │
         v
    ┌─────────┐
    │   S3    │
    └─────────┘
```

**Pros:** Auto-scaling, pay-per-use
**Cons:** Cold starts, vendor lock-in

---

## Future Architecture

### Phase 6: Web UI

```
┌────────────────┐
│   Web UI       │
│ (React + API)  │
└───────┬────────┘
        │
        v
┌────────────────┐
│  Query API     │
│ (Go + HTTP)    │
└───────┬────────┘
        │
        v
┌────────────────┐
│    DuckDB      │
└────────────────┘
```

### Phase 7: Real-Time Queries

```
┌────────────┐
│  Collector │──> S3 (batch)
└────────────┘
      │
      └──────────> Kafka (stream)
                      │
                      v
                 ┌────────────┐
                 │  DuckDB    │
                 │  + Kafka   │
                 │  Connector │
                 └────────────┘
```

### Phase 8: Alerts

```
┌────────────┐
│  Query     │──> Check thresholds
└────────────┘        │
                      v
                 ┌────────────┐
                 │ Alertmanager│
                 └────────────┘
                      │
                      v
                 Slack/PagerDuty
```

---

## Conclusion

openchami-logq is designed for simplicity and cost-effectiveness over performance. It trades real-time queries for cheap storage and zero maintenance.

**Key Architectural Decisions:**
1. ✅ S3 for cheap, scalable storage
2. ✅ NDJSON for schema flexibility
3. ✅ Parquet for query performance
4. ✅ DuckDB for embedded analytics
5. ✅ Streaming pipelines for constant memory

**When to Use:**
- Long-term log retention
- Ad-hoc analysis
- Compliance/auditing
- Cost-sensitive environments

**When NOT to Use:**
- Real-time dashboards
- High-frequency queries
- Sub-second latency requirements
- Complex JOINs across datasets

---

**Questions or feedback?** Open an issue on GitHub!
