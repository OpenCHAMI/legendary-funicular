<!--
SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC

SPDX-License-Identifier: MIT
-->

# Operations Guide

Practical deployment and operations guide for openchami-logq.

## Table of Contents

1. [Deployment](#deployment)
2. [Configuration Management](#configuration-management)
3. [Monitoring](#monitoring)
4. [Troubleshooting](#troubleshooting)
5. [Maintenance](#maintenance)

---

## Deployment

### Architecture Overview

```
┌─────────────┐
│  Collector  │──> S3 Raw Bucket
│  (Vector)   │
└─────────────┘
                    ┌──────────────┐
                    │  Compactor   │──> S3 Compacted Bucket
                    │  (Cron Job)  │
                    └──────────────┘
                                        ┌──────────────┐
                                        │    Query     │
                                        │    (CLI)     │
                                        └──────────────┘
```

### Prerequisites

**Infrastructure:**
- S3-compatible storage (VersityGW, MinIO, or AWS S3)
- Linux server with systemd or Podman
- IAM users with appropriate permissions

**Software:**
- Vector or FluentBit (for log collection)
- Podman or Docker
- Systemd (for service management)

---

## Deployment Option 1: Podman Quadlet

**Best for:** Production deployments with systemd integration

### 1. Install VersityGW

```bash
# Install and start VersityGW
# See: https://github.com/versity/versitygw

systemctl enable versitygw
systemctl start versitygw
```

### 2. Create S3 Buckets

```bash
# Configure AWS CLI
aws configure set aws_access_key_id admin
aws configure set aws_secret_access_key admin-secret
aws configure set default.region us-east-1

# Create buckets
aws s3 mb s3://openchami-logs-raw --endpoint-url=http://localhost:7070
aws s3 mb s3://openchami-logs-daily --endpoint-url=http://localhost:7070
```

### 3. Create IAM Users

Create three users with appropriate permissions:

**log-writer:**
- `s3:PutObject` on raw bucket

**log-compactor:**
- `s3:GetObject`, `s3:ListBucket` on raw bucket
- `s3:PutObject` on compacted bucket
- `s3:DeleteObject` on raw bucket

**log-reader:**
- `s3:GetObject`, `s3:ListBucket` on compacted bucket

### 4. Configure Vector Collector

Create `/etc/vector/vector.yaml`:

```yaml
sources:
  syslog:
    type: syslog
    mode: tcp
    address: 0.0.0.0:514

transforms:
  parse:
    type: remap
    inputs: ["syslog"]
    source: |
      .ts = .timestamp
      .host = .hostname
      .msg = .message
      .level = .severity

sinks:
  s3:
    type: aws_s3
    inputs: ["parse"]
    bucket: openchami-logs-raw
    key_prefix: logs/
    compression: none
    encoding:
      codec: ndjson
    batch:
      max_bytes: 10485760  # 10MB
      timeout_secs: 300     # 5 minutes
    auth:
      access_key_id: "${S3_ACCESS_KEY}"
      secret_access_key: "${S3_SECRET_KEY}"
    endpoint: "http://localhost:7070"
    region: us-east-1
```

Create `/etc/systemd/system/vector.service`:

```ini
[Unit]
Description=Vector Log Collector
After=network.target

[Service]
Type=simple
User=vector
EnvironmentFile=/etc/vector/vector.env
ExecStart=/usr/bin/vector --config /etc/vector/vector.yaml
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

Create `/etc/vector/vector.env`:

```bash
S3_ACCESS_KEY=log-writer-key
S3_SECRET_KEY=log-writer-secret
```

Start Vector:

```bash
systemctl enable vector
systemctl start vector
```

### 5. Configure Compactor (Quadlet)

Create `/etc/containers/systemd/openchami-logq-compactor.container`:

```ini
[Unit]
Description=OpenCHAMI Log Compactor
After=network-online.target

[Container]
Image=ghcr.io/openchami/logq-compactor:latest
Environment=S3_ENDPOINT=http://host.containers.internal:7070
Environment=S3_ACCESS_KEY=log-compactor-key
Environment=S3_SECRET_KEY=log-compactor-secret
Environment=S3_BUCKET_RAW=openchami-logs-raw
Environment=S3_BUCKET_COMPACTED=openchami-logs-daily
Environment=S3_SSL=false
Network=host

[Service]
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

Create timer for daily compaction at 2 AM:

Create `/etc/containers/systemd/openchami-logq-compactor.timer`:

```ini
[Unit]
Description=Run OpenCHAMI Log Compactor Daily

[Timer]
OnCalendar=daily
OnCalendar=*-*-* 02:00:00
Persistent=true

[Install]
WantedBy=timers.target
```

Enable the timer:

```bash
systemctl daemon-reload
systemctl enable openchami-logq-compactor.timer
systemctl start openchami-logq-compactor.timer
```

### 6. Install Query CLI

```bash
# Download binary
curl -sSL https://raw.githubusercontent.com/OpenCHAMI/legendary-funicular/main/installer.bash | bash

# Configure
cat > ~/.openchami-logq.env <<EOF
export S3_ENDPOINT="http://localhost:7070"
export S3_ACCESS_KEY="log-reader-key"
export S3_SECRET_KEY="log-reader-secret"
export S3_SSL="false"
EOF

source ~/.openchami-logq.env

# Test
openchami-logq-query inspect dates
```

---

## Deployment Option 2: Docker Compose

**Best for:** Development, testing, and simple deployments

### Complete Stack

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  # S3 Storage (VersityGW)
  versitygw:
    image: versity/versitygw:latest
    ports:
      - "7070:7070"
    environment:
      - VERSITY_GW_ENDPOINT=:7070
      - VERSITY_GW_ACCESS_KEY=admin
      - VERSITY_GW_SECRET_KEY=admin-secret
    volumes:
      - versity-data:/data
    restart: unless-stopped

  # Log Collector (Vector)
  vector:
    image: timberio/vector:latest
    ports:
      - "514:514/tcp"
    volumes:
      - ./vector.yaml:/etc/vector/vector.yaml:ro
    environment:
      - S3_ENDPOINT=http://versitygw:7070
      - S3_ACCESS_KEY=log-writer
      - S3_SECRET_KEY=log-writer-secret
    depends_on:
      - versitygw
    restart: unless-stopped

  # Compactor (runs daily)
  compactor:
    image: ghcr.io/openchami/logq-compactor:latest
    environment:
      - S3_ENDPOINT=http://versitygw:7070
      - S3_ACCESS_KEY=log-compactor
      - S3_SECRET_KEY=log-compactor-secret
      - S3_BUCKET_RAW=openchami-logs-raw
      - S3_BUCKET_COMPACTED=openchami-logs-daily
      - S3_SSL=false
    depends_on:
      - versitygw
    restart: on-failure
    # Run daily at 2 AM (requires external scheduler like cron or systemd timer)

volumes:
  versity-data:
```

Create `vector.yaml`:

```yaml
sources:
  syslog:
    type: syslog
    mode: tcp
    address: 0.0.0.0:514

transforms:
  parse:
    type: remap
    inputs: ["syslog"]
    source: |
      .ts = .timestamp
      .host = .hostname
      .msg = .message
      .level = .severity

sinks:
  s3:
    type: aws_s3
    inputs: ["parse"]
    bucket: openchami-logs-raw
    key_prefix: logs/
    compression: none
    encoding:
      codec: ndjson
    batch:
      max_bytes: 10485760
      timeout_secs: 300
    auth:
      access_key_id: "${S3_ACCESS_KEY}"
      secret_access_key: "${S3_SECRET_KEY}"
    endpoint: "${S3_ENDPOINT}"
    region: us-east-1
```

Start the stack:

```bash
# Start services
docker-compose up -d

# Create buckets
docker-compose exec versitygw aws s3 mb s3://openchami-logs-raw --endpoint-url=http://localhost:7070
docker-compose exec versitygw aws s3 mb s3://openchami-logs-daily --endpoint-url=http://localhost:7070

# Run compactor manually
docker-compose run --rm compactor

# Query logs
docker run --rm \
  -e S3_ENDPOINT=http://host.docker.internal:7070 \
  -e S3_ACCESS_KEY=log-reader \
  -e S3_SECRET_KEY=log-reader-secret \
  -e S3_SSL=false \
  ghcr.io/openchami/logq-query:latest \
  sql "SELECT COUNT(*) FROM SOURCES"
```

---

## Configuration Management

For configuration details and environment variables, see the [main README](../README.md#configuration).

### IAM Permissions Summary

| User | Bucket | Permissions |
|------|--------|-------------|
| log-writer | raw | `s3:PutObject` |
| log-compactor | raw | `s3:GetObject`, `s3:ListBucket`, `s3:DeleteObject` |
| log-compactor | compacted | `s3:PutObject` |
| log-reader | compacted | `s3:GetObject`, `s3:ListBucket` |

### Secrets Management

**Development:**
- Environment files with `chmod 600` permissions
- Store in `/etc/openchami-logq/*.env`

**Production:**
- Use systemd credential files
- Or external secrets manager (Vault, AWS Secrets Manager)

---

## Monitoring

### Key Metrics to Monitor

**Collector:**
- Logs ingested per second
- S3 upload failures
- Buffer usage

**Compactor:**
- Files processed per run
- Compression ratio
- S3 errors

**Storage:**
- Raw bucket size
- Compacted bucket size
- S3 request errors

### Simple Monitoring Script

```bash
#!/bin/bash
# /usr/local/bin/check-logq-health.sh

# Check raw bucket size
RAW_SIZE=$(aws s3 ls s3://openchami-logs-raw --recursive --endpoint-url=$S3_ENDPOINT | \
  awk '{sum+=$3} END {print sum/1024/1024/1024}')

# Check compacted bucket size
COMPACTED_SIZE=$(aws s3 ls s3://openchami-logs-daily --recursive --endpoint-url=$S3_ENDPOINT | \
  awk '{sum+=$3} END {print sum/1024/1024/1024}')

echo "Raw bucket: ${RAW_SIZE}GB"
echo "Compacted bucket: ${COMPACTED_SIZE}GB"

# Check if compactor ran recently
LAST_COMPACTED=$(aws s3 ls s3://openchami-logs-daily/logs/ --endpoint-url=$S3_ENDPOINT | \
  tail -1 | awk '{print $1}')

echo "Last compaction: $LAST_COMPACTED"
```

### Logging

**Vector logs:**
```bash
journalctl -u vector -f
```

**Compactor logs:**
```bash
# Quadlet
journalctl -u openchami-logq-compactor -f

# Docker Compose
docker-compose logs -f compactor
```

---

## Troubleshooting

### Common Issues

#### Vector Can't Write to S3

**Symptoms:**
- Logs not appearing in raw bucket
- Vector errors about S3 access

**Solutions:**
```bash
# Check credentials
aws s3 ls s3://openchami-logs-raw --endpoint-url=$S3_ENDPOINT

# Check Vector logs
journalctl -u vector -n 50

# Verify bucket exists
aws s3 ls --endpoint-url=$S3_ENDPOINT

# Test write permission
echo "test" | aws s3 cp - s3://openchami-logs-raw/test.txt --endpoint-url=$S3_ENDPOINT
```

#### Compactor Fails

**Symptoms:**
- Raw files accumulating
- No new Parquet files in compacted bucket

**Solutions:**
```bash
# Check compactor logs
journalctl -u openchami-logq-compactor -n 50

# Run manually with verbose output
openchami-logq-compactor --verbose

# Check permissions
aws s3 ls s3://openchami-logs-raw/logs/ --endpoint-url=$S3_ENDPOINT
aws s3 ls s3://openchami-logs-daily/logs/ --endpoint-url=$S3_ENDPOINT
```

#### Query Returns No Results

**Symptoms:**
- Query runs but returns empty results
- `inspect dates` shows no dates

**Solutions:**
```bash
# Check available dates
openchami-logq-query inspect dates --stream logs

# Check if data exists in S3
aws s3 ls s3://openchami-logs-daily/logs/ --endpoint-url=$S3_ENDPOINT

# Verify configuration
openchami-logq-query inspect config

# Try querying raw data
openchami-logq-query sql --scope raw "SELECT COUNT(*) FROM SOURCES"
```

#### S3 Storage Full

**Symptoms:**
- Vector fails to write
- Compactor fails

**Solutions:**
```bash
# Check bucket sizes
aws s3 ls s3://openchami-logs-raw --recursive --endpoint-url=$S3_ENDPOINT | \
  awk '{sum+=$3} END {print "Raw: " sum/1024/1024/1024 "GB"}'

aws s3 ls s3://openchami-logs-daily --recursive --endpoint-url=$S3_ENDPOINT | \
  awk '{sum+=$3} END {print "Compacted: " sum/1024/1024/1024 "GB"}'

# Clean up old raw files (if compaction successful)
# WARNING: Only do this if compacted data exists!
aws s3 rm s3://openchami-logs-raw/logs/ --recursive --endpoint-url=$S3_ENDPOINT --exclude "*" --include "$(date -d '7 days ago' +%Y-%m-%d)*"
```

---

## Maintenance

### Regular Tasks

**Daily (Automated):**
- Compactor runs via timer/cron
- Logs are compacted to Parquet

**Weekly:**
- Check storage usage
- Review error logs
- Verify compaction is working

**Monthly:**
- Review retention policy
- Clean up old compacted data if needed

### Maintenance Script

```bash
#!/bin/bash
# /usr/local/bin/logq-maintenance.sh

# Check storage
echo "=== Storage Usage ==="
aws s3 ls s3://openchami-logs-raw --recursive --endpoint-url=$S3_ENDPOINT | \
  awk '{sum+=$3} END {print "Raw: " sum/1024/1024/1024 "GB"}'
aws s3 ls s3://openchami-logs-daily --recursive --endpoint-url=$S3_ENDPOINT | \
  awk '{sum+=$3} END {print "Compacted: " sum/1024/1024/1024 "GB"}'

# Check recent compaction
echo "=== Recent Compaction ==="
aws s3 ls s3://openchami-logs-daily/logs/ --endpoint-url=$S3_ENDPOINT | tail -5

# Check for errors in last 24h
echo "=== Recent Errors ==="
journalctl -u vector -u openchami-logq-compactor --since "24 hours ago" --priority err

# Test query
echo "=== Test Query ==="
openchami-logq-query sql "SELECT COUNT(*) as count FROM SOURCES" 2>&1 | tail -1
```

### Upgrading

**Query CLI:**
```bash
# Backup current binary
cp /usr/local/bin/openchami-logq-query /usr/local/bin/openchami-logq-query.backup

# Download new version
curl -sSL https://raw.githubusercontent.com/OpenCHAMI/legendary-funicular/main/installer.bash | bash

# Test
openchami-logq-query version
```

**Compactor (Quadlet):**
```bash
# Pull new image
podman pull ghcr.io/openchami/logq-compactor:latest

# Restart service
systemctl restart openchami-logq-compactor
```

**Compactor (Docker Compose):**
```bash
# Pull new image
docker-compose pull compactor

# Restart
docker-compose up -d compactor
```

### Basic Security

**File Permissions:**
```bash
# Secure environment files
chmod 600 /etc/vector/vector.env
chmod 600 /etc/openchami-logq/*.env

# Restrict user access
chown root:root /etc/vector/vector.env
chown root:root /etc/openchami-logq/*.env
```

**Network Security:**
- Use firewall to restrict access to S3 endpoint
- Use SSL/TLS for S3 in production (`S3_SSL=true`)
- Rotate access keys regularly

**IAM Best Practices:**
- Use separate users for each component
- Grant minimum required permissions
- Rotate credentials regularly
- Audit S3 access logs

---

## Additional Resources

- **[Main README](../README.md)** - Installation and basic usage
- **[User Guide](USER_GUIDE.md)** - Advanced SQL queries and DuckDB tips
- **[Architecture](ARCHITECTURE.md)** - System design details
- **[Developer Guide](DEVELOPMENT.md)** - Contributing and development

---

**Need help?** Open an issue on [GitHub](https://github.com/OpenCHAMI/legendary-funicular/issues)!
