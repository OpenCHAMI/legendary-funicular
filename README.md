<!--
SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC

SPDX-License-Identifier: MIT
-->

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

- [x] **Phase 0 - Storage & IAM (current focus)**
  - VersityGW via quadlet
  - Buckets: `openchami-logs-raw`, `openchami-logs-daily`
  - Users: `log-writer`, `log-compactor`, `log-reader`
  - Verify access with awscli

- [x] **Phase 1 - Syslog -> Raw NDJSON**
  - Vector or fluent-bit
  - Syslog in -> hourly NDJSON in S3 **changed: hourly -> minutely to ensure proper disk caching**
  - Never drop logs; schema changes must not break ingestion

- [x] **Phase 2 - CloudEvents -> Raw NDJSON**
  - Accept CloudEvents
  - Store alongside logs, payload preserved

- [x] **Phase 3 - Compaction (DuckDB)**
  - Daily job converts NDJSON → Parquet
  - Best-effort field extraction
  - Parquet readable directly from S3

- [x] **Phase 4 - Query CLI (Go + DuckDB)**
  - CLI: `openchami-logq --date ... --kind logs|events --sql "..."`
  - Queries Parquet directly from VersityGW

- [x] **Phase 5 — Domain queries**
  - Canned OpenCHAMI-relevant queries (errors by service, trace walk, xname
    grouping, etc.)
