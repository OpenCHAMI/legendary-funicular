<!--
SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC

SPDX-License-Identifier: MIT
-->

# Operations Guide

Practical deployment and operations guide for openchami-logq.

## Table of Contents

1. [Deployment](#deployment)
2. [Configuration](#configuration)
3. [Monitoring](#monitoring)
4. [Troubleshooting](#troubleshooting)
5. [Maintenance](#maintenance)

---

## Deployment

### Prerequisites

**Infrastructure:**
- S3-compatible storage ([VersityGW](https://github.com/versity/versitygw), [MinIO](https://min.io/), or AWS S3)
- Linux server with systemd or Docker
- IAM users with appropriate permissions (see [IAM Permissions](#iam-permissions))

**Software:**
- [Vector](https://vector.dev/) or [FluentBit](https://fluentbit.io/) for log collection
- Podman/Docker for containers
- Systemd for service management (Quadlet option)

### Deployment Options

Choose the deployment method that fits your environment:

1. **[Podman Quadlet](#option-1-podman-quadlet)** - Production deployments with systemd integration
2. **[Docker Compose](#option-2-docker-compose)** - Development, testing, and simple deployments

---

## Option 1: Podman Quadlet

**Best for:** Production deployments with systemd integration

### 1. Install S3 Storage

Install VersityGW (or use MinIO/AWS S3):

```bash
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

Create three users with minimal permissions (see [IAM Permissions](#iam-permissions) table).

### 4. Configure Vector Collector

**Install Vector:**
```bash
# See: https://vector.dev/docs/setup/installation/
```

**Copy example configuration:**
```bash
# Copy Vector config
sudo cp deploy/vector/vector.yaml /etc/vector/vector.yaml

# Copy systemd service
sudo cp deploy/systemd/vector.service /etc/systemd/system/vector.service

# Create environment file
sudo bash -c 'cat > /etc/vector/vector.env <<EOF
S3_ENDPOINT=http://localhost:7070
S3_ACCESS_KEY=log-writer-key
S3_SECRET_KEY=log-writer-secret
EOF'

sudo chmod 600 /etc/vector/vector.env
```

**Start Vector:**
```bash
systemctl daemon-reload
systemctl enable vector
systemctl start vector
systemctl status vector
```

**Example files:**
- [`deploy/vector/vector.yaml`](../deploy/vector/vector.yaml) - Vector configuration
- [`deploy/systemd/vector.service`](../deploy/systemd/vector.service) - Systemd unit

### 5. Configure Compactor

**Copy Quadlet configuration:**
```bash
# Copy Quadlet files
sudo cp deploy/systemd/openchami-logq-compactor.container \
  /etc/containers/systemd/openchami-logq-compactor.container

sudo cp deploy/systemd/openchami-logq-compactor.timer \
  /etc/containers/systemd/openchami-logq-compactor.timer

# Edit to set your credentials
sudo nano /etc/containers/systemd/openchami-logq-compactor.container
```

**Enable timer:**
```bash
systemctl daemon-reload
systemctl enable openchami-logq-compactor.timer
systemctl start openchami-logq-compactor.timer
systemctl status openchami-logq-compactor.timer
```

**Example files:**
- [`deploy/systemd/openchami-logq-compactor.container`](../deploy/systemd/openchami-logq-compactor.container) - Quadlet container
- [`deploy/systemd/openchami-logq-compactor.timer`](../deploy/systemd/openchami-logq-compactor.timer) - Systemd timer

### 6. Install Query CLI

```bash
# Download and install
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

## Option 2: Docker Compose

**Best for:** Development, testing, and simple deployments

### Quick Start

```bash
# Copy example docker-compose.yml
cp deploy/docker/docker-compose.yml .
cp deploy/vector/vector.yaml .

# Start services
docker-compose up -d

# Create buckets (wait ~10s for VersityGW to start)
sleep 10
docker-compose exec versitygw \
  aws s3 mb s3://openchami-logs-raw --endpoint-url=http://localhost:7070
docker-compose exec versitygw \
  aws s3 mb s3://openchami-logs-daily --endpoint-url=http://localhost:7070

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

**Example file:**
- [`deploy/docker/docker-compose.yml`](../deploy/docker/docker-compose.yml) - Complete stack

### Scheduling Compactor

Docker Compose doesn't have built-in scheduling. Use one of these options:

**Option A: Cron**
```bash
# Add to crontab
0 2 * * * cd /path/to/project && docker-compose run --rm compactor
```

**Option B: Systemd Timer**
```bash
# Create /etc/systemd/system/logq-compactor.timer
# See deploy/systemd/openchami-logq-compactor.timer for example
```

---

## Configuration

### Environment Variables

For complete configuration reference, see [README Configuration section](../README.md#configuration).

**Key variables:**
- `S3_ENDPOINT` - S3 endpoint URL
- `S3_ACCESS_KEY` - IAM user access key
- `S3_SECRET_KEY` - IAM user secret key
- `S3_BUCKET_RAW` - Raw logs bucket name
- `S3_BUCKET_COMPACTED` - Compacted logs bucket name
- `S3_SSL` - Enable SSL/TLS (`true` or `false`)

### IAM Permissions

| User | Bucket | Permissions |
|------|--------|-------------|
| log-writer | raw | `s3:PutObject` |
| log-compactor | raw | `s3:GetObject`, `s3:ListBucket`, `s3:DeleteObject` |
| log-compactor | compacted | `s3:PutObject` |
| log-reader | compacted | `s3:GetObject`, `s3:ListBucket` |

**Why separate users?**
- Principle of least privilege
- Limits blast radius if credentials compromised
- Easier to audit access

### Secrets Management

**Development:**
```bash
# Use environment files with restricted permissions
chmod 600 /etc/vector/vector.env
chmod 600 ~/.openchami-logq.env
```

**Production:**
- Use [systemd credential files](https://systemd.io/CREDENTIALS/)
- Or external secrets manager ([Vault](https://www.vaultproject.io/), AWS Secrets Manager)
- Rotate credentials regularly

---

## Monitoring

### Key Metrics

**Collector (Vector):**
- Logs ingested per second
- S3 upload failures
- Buffer usage

**Compactor:**
- Files processed per run
- Compression ratio (typically 10:1)
- S3 errors

**Storage:**
- Raw bucket size (should shrink after compaction)
- Compacted bucket size (grows steadily)
- S3 request errors

### Health Check Script

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

# Check last compaction date
LAST_COMPACTED=$(aws s3 ls s3://openchami-logs-daily/logs/ --endpoint-url=$S3_ENDPOINT | \
  tail -1 | awk '{print $1}')

echo "Last compaction: $LAST_COMPACTED"
```

### View Logs

**Vector logs:**
```bash
journalctl -u vector -f
```

**Compactor logs (Quadlet):**
```bash
journalctl -u openchami-logq-compactor -f
```

**Compactor logs (Docker Compose):**
```bash
docker-compose logs -f compactor
```

---

## Troubleshooting

### Vector Can't Write to S3

**Symptoms:** Logs not appearing in raw bucket, Vector errors about S3 access

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

### Compactor Fails

**Symptoms:** Raw files accumulating, no new Parquet files

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

### Query Returns No Results

**Symptoms:** Query runs but returns empty, `inspect dates` shows no dates

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

### S3 Storage Full

**Symptoms:** Vector/Compactor fails to write

**Solutions:**
```bash
# Check bucket sizes
aws s3 ls s3://openchami-logs-raw --recursive --endpoint-url=$S3_ENDPOINT | \
  awk '{sum+=$3} END {print "Raw: " sum/1024/1024/1024 "GB"}'

aws s3 ls s3://openchami-logs-daily --recursive --endpoint-url=$S3_ENDPOINT | \
  awk '{sum+=$3} END {print "Compacted: " sum/1024/1024/1024 "GB"}'

# Clean up old raw files (WARNING: Only if compacted data exists!)
aws s3 rm s3://openchami-logs-raw/logs/ --recursive \
  --endpoint-url=$S3_ENDPOINT \
  --exclude "*" \
  --include "$(date -d '7 days ago' +%Y-%m-%d)*"
```

---

## Maintenance

### Regular Tasks

**Daily (Automated):**
- Compactor runs via timer/cron
- Logs compacted to Parquet

**Weekly:**
- Check storage usage
- Review error logs
- Verify compaction is working

**Monthly:**
- Review retention policy
- Clean up old data if needed

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
podman pull ghcr.io/openchami/logq-compactor:latest
systemctl restart openchami-logq-compactor
```

**Compactor (Docker Compose):**
```bash
docker-compose pull compactor
docker-compose up -d compactor
```

### Security Best Practices

**File Permissions:**
```bash
chmod 600 /etc/vector/vector.env
chmod 600 /etc/openchami-logq/*.env
chown root:root /etc/vector/vector.env
```

**Network Security:**
- Use firewall to restrict S3 endpoint access
- Enable SSL/TLS in production (`S3_SSL=true`)
- Use VPN or private networks for S3 traffic

**IAM Best Practices:**
- Use separate users for each component
- Grant minimum required permissions
- Rotate credentials regularly
- Audit S3 access logs

---

## Related Documentation

- **[Main README](../README.md)** - Installation and configuration
- **[User Guide](USER_GUIDE.md)** - Advanced queries and SQL patterns
- **[Architecture](ARCHITECTURE.md)** - System design and technical details
- **[Developer Guide](DEVELOPMENT.md)** - Contributing and development setup

---

**Need help?** Open an issue on [GitHub](https://github.com/OpenCHAMI/legendary-funicular/issues)!
