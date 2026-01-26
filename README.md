# openchami-logq

Lightweight log lake + DuckDB query tool for OpenCHAMI.

**Goal:** preserve all logs/events safely (schema-flexible), store them cheaply,
and make them easy to query with minimal operational overhead. Not
ELK/Kafka/ClickHouse.

Stack:

- VersityGW
- NDJSON
- Parquet
- DuckDB

## Development phases (checklist)

- [ ] **Phase 0 - Storage & IAM (current focus)**
  - VersityGW via quadlet
  - Buckets: `openchami-logs-raw`, `openchami-logs-daily`
  - Users: `log-writer`, `log-compactor`, `log-reader`
  - Verify access with awscli

- [ ] **Phase 1 - Syslog -> Raw NDJSON**
  - Vector or fluent-bit
  - Syslog in -> hourly NDJSON in S3
  - Never drop logs; schema changes must not break ingestion

- [ ] **Phase 2 - CloudEvents -> Raw NDJSON**
  - Accept CloudEvents
  - Store alongside logs, payload preserved

- [ ] **Phase 3 - Compaction (DuckDB)**
  - Daily job converts NDJSON → Parquet
  - Best-effort field extraction
  - Parquet readable directly from S3

- [ ] **Phase 4 - Query CLI (Go + DuckDB)**
  - CLI: `openchami-logq --date ... --kind logs|events --sql "..."`
  - Queries Parquet directly from VersityGW

- [ ] **Phase 5 — Domain queries**
  - Canned OpenCHAMI-relevant queries (errors by service, trace walk, xname
    grouping, etc.)
