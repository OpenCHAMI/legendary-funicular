<!--
SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC

SPDX-License-Identifier: MIT
-->

# Operations Guide

**Project:** openchami-logq
**Version:** 1.0
**Last Updated:** June 10, 2026

---

## Table of Contents

1. [Deployment](#deployment)
2. [Configuration Management](#configuration-management)
3. [Monitoring](#monitoring)
4. [Backup and Recovery](#backup-and-recovery)
5. [Scaling](#scaling)
6. [Security](#security)
7. [Troubleshooting](#troubleshooting)
8. [Maintenance](#maintenance)
9. [Performance Tuning](#performance-tuning)
10. [Disaster Recovery](#disaster-recovery)

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
- Linux server (for collector and compactor)
- IAM users with appropriate permissions

**Software:**
- Vector or FluentBit (for log collection)
- Systemd (for service management)
- Cron (for scheduling compactor)

---

### Deployment Option 1: Single Server

**Best for:** Small deployments, testing, single-cluster setups

#### 1. Install VersityGW (S3-compatible storage)

```bash
# Install VersityGW
# See: https://github.com/versity/versitygw

# Start VersityGW
systemctl enable versitygw
systemctl start versitygw
```

#### 2. Create S3 Buckets

```bash
# Configure AWS CLI
aws configure set aws_access_key_id admin
aws configure set aws_secret_access_key admin-secret
aws configure set default.region us-east-1

# Create buckets
aws s3 mb s3://openchami-logs-raw --endpoint-url=http://localhost:7070
aws s3 mb s3://openchami-logs-daily --endpoint-url=http://localhost:7070
```

#### 3. Create IAM Users

```bash
# Create log-writer user (write-only to raw bucket)
# Create log-compactor user (read raw, write/delete compacted)
# Create log-reader user (read-only compacted bucket)

# See VersityGW documentation for IAM user creation
```

#### 4. Install openchami-logq

```bash
# Download installer
curl -sSL https://raw.githubusercontent.com/OpenCHAMI/legendary-funicular/main/installer.bash -o installer.bash

# Run installer
bash installer.bash

# Verify installation
openchami-logq-query version
openchami-logq-compactor --help
```

#### 5. Configure Collector (Vector)

Create `/etc/vector/vector.yaml`:

```yaml
sources:
  syslog:
    type: syslog
    address: 0.0.0.0:514
    mode: tcp

  cloudevents:
    type: http_server
    address: 0.0.0.0:8080
    encoding: json

transforms:
  parse_logs:
    type: remap
    inputs:
      - syslog
      - cloudevents
    source: |
      .ts = now()
      .data = .

sinks:
  s3_raw:
    type: aws_s3
    inputs:
      - parse_logs
    region: us-east-1
    endpoint: http://localhost:7070
    bucket: openchami-logs-raw
    key_prefix: "logs/{{ timestamp | date('%Y-%m-%d') }}/"
    filename_time_format: "%H-%M-%S"
    filename_extension: ndjson.zst
    compression: zstd
    encoding:
      codec: ndjson
    batch:
      max_bytes: 10485760  # 10MB
      timeout_secs: 60
    auth:
      access_key_id: "${S3_ACCESS_KEY}"
      secret_access_key: "${S3_SECRET_KEY}"
```

Start Vector:

```bash
# Create environment file
cat > /etc/vector/env <<EOF
S3_ACCESS_KEY=log-writer-key
S3_SECRET_KEY=log-writer-secret
EOF

# Start Vector
systemctl enable vector
systemctl start vector
```

#### 6. Configure Compactor (Systemd)

Create `/etc/systemd/system/openchami-logq-compactor.service`:

```ini
[Unit]
Description=OpenCHAMI Log Compactor
After=network.target

[Service]
Type=oneshot
User=logq
EnvironmentFile=/etc/openchami-logq/compactor.env
ExecStart=/usr/local/bin/openchami-logq-compactor
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

Create `/etc/systemd/system/openchami-logq-compactor.timer`:

```ini
[Unit]
Description=Run OpenCHAMI Log Compactor daily

[Timer]
OnCalendar=daily
OnCalendar=02:00
Persistent=true

[Install]
WantedBy=timers.target
```

Create `/etc/openchami-logq/compactor.env`:

```bash
S3_ENDPOINT=http://localhost:7070
S3_REGION=us-east-1
S3_ACCESS_KEY=log-compactor-key
S3_SECRET_KEY=log-compactor-secret
S3_BUCKET_SOURCE=openchami-logs-raw
S3_BUCKET_SINK=openchami-logs-daily
S3_SSL=false
```

Enable and start:

```bash
# Create user
useradd -r -s /bin/false logq

# Set permissions
chmod 600 /etc/openchami-logq/compactor.env
chown logq:logq /etc/openchami-logq/compactor.env

# Enable timer
systemctl enable openchami-logq-compactor.timer
systemctl start openchami-logq-compactor.timer

# Test compactor manually
systemctl start openchami-logq-compactor
journalctl -u openchami-logq-compactor -f
```

#### 7. Configure Query Access

Create `/etc/profile.d/openchami-logq.sh`:

```bash
export S3_ENDPOINT="http://localhost:7070"
export S3_REGION="us-east-1"
export S3_ACCESS_KEY="log-reader-key"
export S3_SECRET_KEY="log-reader-secret"
export S3_BUCKET_COMPACTED="openchami-logs-daily"
export S3_SSL="false"
```

Test query:

```bash
source /etc/profile.d/openchami-logq.sh
openchami-logq-query sql "SELECT COUNT(*) FROM SOURCES"
```

---

### Deployment Option 2: Kubernetes

**Best for:** Large deployments, multi-cluster setups, cloud environments

#### 1. Install S3 Storage (MinIO)

```yaml
# minio.yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: minio-pvc
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 100Gi
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: minio
spec:
  replicas: 1
  selector:
    matchLabels:
      app: minio
  template:
    metadata:
      labels:
        app: minio
    spec:
      containers:
      - name: minio
        image: minio/minio:latest
        args:
          - server
          - /data
          - --console-address
          - :9001
        env:
        - name: MINIO_ROOT_USER
          value: admin
        - name: MINIO_ROOT_PASSWORD
          value: admin-secret
        ports:
        - containerPort: 9000
        - containerPort: 9001
        volumeMounts:
        - name: data
          mountPath: /data
      volumes:
      - name: data
        persistentVolumeClaim:
          claimName: minio-pvc
---
apiVersion: v1
kind: Service
metadata:
  name: minio
spec:
  selector:
    app: minio
  ports:
  - name: api
    port: 9000
    targetPort: 9000
  - name: console
    port: 9001
    targetPort: 9001
```

Apply:

```bash
kubectl apply -f minio.yaml
```

#### 2. Create Secrets

```bash
# Create S3 credentials secret
kubectl create secret generic openchami-logq-s3 \
  --from-literal=endpoint=http://minio:9000 \
  --from-literal=region=us-east-1 \
  --from-literal=writer-key=log-writer-key \
  --from-literal=writer-secret=log-writer-secret \
  --from-literal=compactor-key=log-compactor-key \
  --from-literal=compactor-secret=log-compactor-secret \
  --from-literal=reader-key=log-reader-key \
  --from-literal=reader-secret=log-reader-secret
```

#### 3. Deploy Collector (Vector)

```yaml
# vector.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: vector-config
data:
  vector.yaml: |
    sources:
      syslog:
        type: syslog
        address: 0.0.0.0:514
        mode: tcp
    sinks:
      s3_raw:
        type: aws_s3
        inputs:
          - syslog
        region: us-east-1
        endpoint: http://minio:9000
        bucket: openchami-logs-raw
        key_prefix: "logs/{{ timestamp | date('%Y-%m-%d') }}/"
        compression: zstd
        encoding:
          codec: ndjson
        auth:
          access_key_id: "${S3_ACCESS_KEY}"
          secret_access_key: "${S3_SECRET_KEY}"
---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: vector
spec:
  selector:
    matchLabels:
      app: vector
  template:
    metadata:
      labels:
        app: vector
    spec:
      containers:
      - name: vector
        image: timberio/vector:latest
        env:
        - name: S3_ACCESS_KEY
          valueFrom:
            secretKeyRef:
              name: openchami-logq-s3
              key: writer-key
        - name: S3_SECRET_KEY
          valueFrom:
            secretKeyRef:
              name: openchami-logq-s3
              key: writer-secret
        volumeMounts:
        - name: config
          mountPath: /etc/vector
      volumes:
      - name: config
        configMap:
          name: vector-config
---
apiVersion: v1
kind: Service
metadata:
  name: vector
spec:
  selector:
    app: vector
  ports:
  - name: syslog
    port: 514
    targetPort: 514
    protocol: TCP
```

Apply:

```bash
kubectl apply -f vector.yaml
```

#### 4. Deploy Compactor (CronJob)

```yaml
# compactor.yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: openchami-logq-compactor
spec:
  schedule: "0 2 * * *"  # Daily at 2 AM
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: compactor
            image: ghcr.io/openchami/logq-compactor:latest
            env:
            - name: S3_ENDPOINT
              valueFrom:
                secretKeyRef:
                  name: openchami-logq-s3
                  key: endpoint
            - name: S3_REGION
              valueFrom:
                secretKeyRef:
                  name: openchami-logq-s3
                  key: region
            - name: S3_ACCESS_KEY
              valueFrom:
                secretKeyRef:
                  name: openchami-logq-s3
                  key: compactor-key
            - name: S3_SECRET_KEY
              valueFrom:
                secretKeyRef:
                  name: openchami-logq-s3
                  key: compactor-secret
            - name: S3_BUCKET_SOURCE
              value: openchami-logs-raw
            - name: S3_BUCKET_SINK
              value: openchami-logs-daily
            - name: S3_SSL
              value: "false"
          restartPolicy: OnFailure
```

Apply:

```bash
kubectl apply -f compactor.yaml

# Test manually
kubectl create job --from=cronjob/openchami-logq-compactor compactor-test
kubectl logs -f job/compactor-test
```

#### 5. Deploy Query (Job)

```yaml
# query-job.yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: openchami-logq-query
spec:
  template:
    spec:
      containers:
      - name: query
        image: ghcr.io/openchami/logq-query:latest
        args:
          - sql
          - "SELECT COUNT(*) FROM SOURCES"
        env:
        - name: S3_ENDPOINT
          valueFrom:
            secretKeyRef:
              name: openchami-logq-s3
              key: endpoint
        - name: S3_REGION
          valueFrom:
            secretKeyRef:
              name: openchami-logq-s3
              key: region
        - name: S3_ACCESS_KEY
          valueFrom:
            secretKeyRef:
              name: openchami-logq-s3
              key: reader-key
        - name: S3_SECRET_KEY
          valueFrom:
            secretKeyRef:
              name: openchami-logq-s3
              key: reader-secret
        - name: S3_BUCKET_COMPACTED
          value: openchami-logs-daily
      restartPolicy: Never
```

Run query:

```bash
kubectl apply -f query-job.yaml
kubectl logs job/openchami-logq-query
```

---

### Deployment Option 3: Docker Compose

**Best for:** Development, testing, small deployments

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  minio:
    image: minio/minio:latest
    command: server /data --console-address :9001
    environment:
      MINIO_ROOT_USER: admin
      MINIO_ROOT_PASSWORD: admin-secret
    ports:
      - "9000:9000"
      - "9001:9001"
    volumes:
      - minio-data:/data

  vector:
    image: timberio/vector:latest
    depends_on:
      - minio
    environment:
      S3_ACCESS_KEY: log-writer-key
      S3_SECRET_KEY: log-writer-secret
    volumes:
      - ./vector.yaml:/etc/vector/vector.yaml:ro
    ports:
      - "514:514"
      - "8080:8080"

  compactor:
    image: ghcr.io/openchami/logq-compactor:latest
    depends_on:
      - minio
    environment:
      S3_ENDPOINT: http://minio:9000
      S3_REGION: us-east-1
      S3_ACCESS_KEY: log-compactor-key
      S3_SECRET_KEY: log-compactor-secret
      S3_BUCKET_SOURCE: openchami-logs-raw
      S3_BUCKET_SINK: openchami-logs-daily
      S3_SSL: "false"
    # Run daily at 2 AM (requires cron in container or external scheduler)

volumes:
  minio-data:
```

Start:

```bash
docker-compose up -d
```

---

## Configuration Management

### Environment Variables

**Collector (Vector):**
```bash
S3_ACCESS_KEY=log-writer-key
S3_SECRET_KEY=log-writer-secret
```

**Compactor:**
```bash
S3_ENDPOINT=http://localhost:7070
S3_REGION=us-east-1
S3_ACCESS_KEY=log-compactor-key
S3_SECRET_KEY=log-compactor-secret
S3_BUCKET_SOURCE=openchami-logs-raw
S3_BUCKET_SINK=openchami-logs-daily
S3_SSL=false
```

**Query:**
```bash
S3_ENDPOINT=http://localhost:7070
S3_REGION=us-east-1
S3_ACCESS_KEY=log-reader-key
S3_SECRET_KEY=log-reader-secret
S3_BUCKET_COMPACTED=openchami-logs-daily
S3_SSL=false
```

### Secrets Management

**Development:**
- Environment files (`/etc/openchami-logq/*.env`)
- File permissions: `chmod 600`

**Production:**
- AWS Secrets Manager
- HashiCorp Vault
- Kubernetes Secrets

**Example (AWS Secrets Manager):**

```bash
# Store secret
aws secretsmanager create-secret \
  --name openchami-logq/s3-credentials \
  --secret-string '{"access_key":"...","secret_key":"..."}'

# Retrieve in script
SECRET=$(aws secretsmanager get-secret-value \
  --secret-id openchami-logq/s3-credentials \
  --query SecretString --output text)

export S3_ACCESS_KEY=$(echo $SECRET | jq -r '.access_key')
export S3_SECRET_KEY=$(echo $SECRET | jq -r '.secret_key')
```

---

## Monitoring

### Metrics to Monitor

**Collector:**
- Logs ingested per second
- S3 upload failures
- Buffer usage
- Memory usage

**Compactor:**
- Compaction duration
- Files processed
- Bytes written
- Errors

**Storage:**
- S3 bucket size
- S3 request rate
- S3 errors

**Query:**
- Query latency
- Query errors
- Concurrent queries

### Logging

**Collector (Vector):**
```bash
journalctl -u vector -f
```

**Compactor:**
```bash
journalctl -u openchami-logq-compactor -f
```

**Query:**
```bash
# CLI output is logged to stdout/stderr
openchami-logq-query sql "SELECT * FROM SOURCES" 2>&1 | tee query.log
```

### Alerting

**Compactor Failures:**
```bash
# Check if compactor failed in last 24 hours
systemctl status openchami-logq-compactor

# Alert if failed
if systemctl is-failed openchami-logq-compactor; then
    echo "Compactor failed!" | mail -s "Alert: Compactor Failed" ops@example.com
fi
```

**S3 Storage Usage:**
```bash
# Check bucket size
aws s3 ls s3://openchami-logs-daily --recursive --summarize \
  --endpoint-url=$S3_ENDPOINT | grep "Total Size"

# Alert if over threshold
SIZE=$(aws s3 ls s3://openchami-logs-daily --recursive --summarize \
  --endpoint-url=$S3_ENDPOINT | grep "Total Size" | awk '{print $3}')

if [ $SIZE -gt 1000000000000 ]; then  # 1TB
    echo "Storage usage: $SIZE bytes" | mail -s "Alert: High Storage" ops@example.com
fi
```

---

## Backup and Recovery

### S3 Bucket Replication

**Enable versioning:**
```bash
aws s3api put-bucket-versioning \
  --bucket openchami-logs-daily \
  --versioning-configuration Status=Enabled \
  --endpoint-url=$S3_ENDPOINT
```

**Cross-region replication:**
```bash
# Configure replication to backup bucket
aws s3api put-bucket-replication \
  --bucket openchami-logs-daily \
  --replication-configuration file://replication.json \
  --endpoint-url=$S3_ENDPOINT
```

### Backup Strategy

**What to backup:**
- ✅ S3 compacted bucket (Parquet files)
- ⚠️ S3 raw bucket (optional, can be deleted after compaction)
- ✅ Configuration files
- ✅ IAM credentials (securely)

**Backup schedule:**
- Daily: Incremental backup of new Parquet files
- Weekly: Full backup of all data
- Monthly: Archive to cold storage (Glacier)

**Retention:**
- Hot storage: 30 days
- Warm storage: 90 days
- Cold storage: 1 year+

### Recovery Procedures

**Scenario 1: Accidental deletion of Parquet files**

```bash
# Restore from S3 versioning
aws s3api list-object-versions \
  --bucket openchami-logs-daily \
  --prefix logs/date=2026-06-10/ \
  --endpoint-url=$S3_ENDPOINT

# Restore specific version
aws s3api restore-object \
  --bucket openchami-logs-daily \
  --key logs/date=2026-06-10/data.parquet \
  --version-id <version-id> \
  --endpoint-url=$S3_ENDPOINT
```

**Scenario 2: Compactor failure (data loss)**

```bash
# Re-run compactor for specific date
export COMPACTOR_DATE=2026-06-10
systemctl start openchami-logq-compactor
```

**Scenario 3: S3 storage failure**

```bash
# Restore from backup bucket
aws s3 sync s3://openchami-logs-daily-backup s3://openchami-logs-daily \
  --endpoint-url=$S3_ENDPOINT
```

---

## Scaling

### Horizontal Scaling

**Collector:**
- Add more Vector instances
- Use load balancer for syslog input
- Each instance writes independently to S3

**Compactor:**
- Run multiple compactors for different date ranges
- Use distributed locking (e.g., Redis) to prevent conflicts

**Query:**
- Stateless (unlimited concurrent queries)
- No coordination needed

### Vertical Scaling

**Collector:**
- Increase memory for larger buffers
- Increase CPU for compression

**Compactor:**
- Increase memory for larger batches
- Increase CPU for faster parsing

**Storage:**
- S3 scales automatically (no tuning needed)

### Performance Tuning

**Collector (Vector):**
```yaml
sinks:
  s3_raw:
    batch:
      max_bytes: 10485760  # 10MB (increase for better throughput)
      timeout_secs: 60     # Decrease for lower latency
    compression: zstd      # zstd (best) or gzip
```

**Compactor:**
```bash
# Run in parallel for multiple streams
openchami-logq-compactor --stream logs &
openchami-logq-compactor --stream events &
wait
```

**Query:**
```bash
# Use WHERE to filter early (partition pruning)
openchami-logq-query sql \
  "SELECT * FROM SOURCES WHERE ts >= '2026-06-10'"

# Select only needed columns (column pruning)
openchami-logq-query sql \
  "SELECT ts, host, msg FROM SOURCES"
```

---

## Security

### IAM Best Practices

**Principle of least privilege:**
- `log-writer`: Write-only to raw bucket
- `log-compactor`: Read raw, write/delete compacted
- `log-reader`: Read-only compacted bucket

**Rotate credentials regularly:**
```bash
# Rotate every 90 days
aws iam create-access-key --user-name log-writer
# Update configuration
# Delete old key
aws iam delete-access-key --user-name log-writer --access-key-id <old-key>
```

### Network Security

**Firewall rules:**
```bash
# Allow syslog from trusted sources only
iptables -A INPUT -p tcp --dport 514 -s 10.0.0.0/8 -j ACCEPT
iptables -A INPUT -p tcp --dport 514 -j DROP
```

**TLS/SSL:**
```bash
# Enable SSL for S3
export S3_SSL=true
export S3_ENDPOINT=https://s3.example.com
```

### Audit Logging

**Enable S3 access logging:**
```bash
aws s3api put-bucket-logging \
  --bucket openchami-logs-daily \
  --bucket-logging-status file://logging.json \
  --endpoint-url=$S3_ENDPOINT
```

**Monitor access:**
```bash
# Check who accessed data
aws s3api get-bucket-logging \
  --bucket openchami-logs-daily \
  --endpoint-url=$S3_ENDPOINT
```

---

## Troubleshooting

### Common Issues

**Issue: Compactor fails with "S3 access denied"**

**Diagnosis:**
```bash
# Check credentials
aws s3 ls s3://openchami-logs-raw \
  --endpoint-url=$S3_ENDPOINT

# Check IAM policy
aws iam get-user-policy --user-name log-compactor --policy-name s3-access
```

**Solution:**
- Verify credentials in `/etc/openchami-logq/compactor.env`
- Check IAM policy includes `s3:GetObject`, `s3:PutObject`, `s3:DeleteObject`

**Issue: Query returns no results**

**Diagnosis:**
```bash
# Check available dates
openchami-logq-query inspect dates

# Check S3 bucket
aws s3 ls s3://openchami-logs-daily/logs/ \
  --endpoint-url=$S3_ENDPOINT
```

**Solution:**
- Run compactor to generate Parquet files
- Check date range in query

**Issue: Collector not sending logs**

**Diagnosis:**
```bash
# Check Vector status
systemctl status vector

# Check Vector logs
journalctl -u vector -f

# Test S3 access
aws s3 cp test.txt s3://openchami-logs-raw/ \
  --endpoint-url=$S3_ENDPOINT
```

**Solution:**
- Check Vector configuration (`/etc/vector/vector.yaml`)
- Verify S3 credentials
- Check network connectivity

---

## Maintenance

### Regular Tasks

**Daily:**
- Check compactor logs
- Monitor storage usage
- Check for errors

**Weekly:**
- Review query performance
- Check backup status
- Update dependencies

**Monthly:**
- Rotate credentials
- Review IAM policies
- Archive old data to cold storage

**Quarterly:**
- Update openchami-logq to latest version
- Review and optimize queries
- Capacity planning

### Upgrading

**Upgrade process:**

1. **Check release notes**
   ```bash
   # See CHANGELOG.md for breaking changes
   curl -s https://api.github.com/repos/OpenCHAMI/legendary-funicular/releases/latest
   ```

2. **Backup configuration**
   ```bash
   cp -r /etc/openchami-logq /etc/openchami-logq.backup
   ```

3. **Stop services**
   ```bash
   systemctl stop openchami-logq-compactor.timer
   systemctl stop vector
   ```

4. **Install new version**
   ```bash
   curl -sSL https://raw.githubusercontent.com/OpenCHAMI/legendary-funicular/main/installer.bash | bash
   ```

5. **Test**
   ```bash
   openchami-logq-query version
   openchami-logq-compactor --help
   ```

6. **Start services**
   ```bash
   systemctl start vector
   systemctl start openchami-logq-compactor.timer
   ```

7. **Verify**
   ```bash
   systemctl status vector
   systemctl status openchami-logq-compactor.timer
   ```

---

## Performance Tuning

### Collector Tuning

**Increase throughput:**
```yaml
# vector.yaml
sinks:
  s3_raw:
    batch:
      max_bytes: 52428800  # 50MB (up from 10MB)
      timeout_secs: 300    # 5 minutes (up from 60s)
```

**Reduce latency:**
```yaml
# vector.yaml
sinks:
  s3_raw:
    batch:
      max_bytes: 1048576   # 1MB (down from 10MB)
      timeout_secs: 10     # 10 seconds (down from 60s)
```

### Compactor Tuning

**Faster compaction:**
```bash
# Run in parallel
openchami-logq-compactor --stream logs &
openchami-logq-compactor --stream events &
wait
```

**Lower memory usage:**
```bash
# Process smaller batches (future feature)
openchami-logq-compactor --batch-size 1000
```

### Query Tuning

**Optimize queries:**
```sql
-- Good: Filter early
SELECT * FROM SOURCES WHERE ts >= '2026-06-10' AND level = 'ERROR'

-- Bad: Filter late
SELECT * FROM (SELECT * FROM SOURCES) WHERE ts >= '2026-06-10'
```

**Use appropriate scope:**
```bash
# Query recent data (raw)
openchami-logq-query sql --scope raw \
  "SELECT * FROM SOURCES WHERE ts > NOW() - INTERVAL 1 HOUR"

# Query historical data (compacted)
openchami-logq-query sql --scope compacted \
  "SELECT * FROM SOURCES WHERE ts >= '2026-06-01'"
```

---

## Disaster Recovery

### Disaster Scenarios

**Scenario 1: Complete S3 storage loss**

**Recovery:**
1. Restore S3 buckets from backup
2. Verify data integrity
3. Resume normal operations

**Scenario 2: Compactor data corruption**

**Recovery:**
1. Delete corrupted Parquet files
2. Re-run compactor for affected dates
3. Verify data integrity

**Scenario 3: Configuration loss**

**Recovery:**
1. Restore configuration from backup
2. Recreate IAM users if needed
3. Test services

### Disaster Recovery Plan

**RTO (Recovery Time Objective):** 4 hours
**RPO (Recovery Point Objective):** 24 hours

**Recovery steps:**
1. Assess damage (30 minutes)
2. Restore S3 buckets from backup (2 hours)
3. Reconfigure services (1 hour)
4. Verify and test (30 minutes)

---

## Additional Resources

- **[Architecture](ARCHITECTURE.md)** - System design
- **[User Guide](USER_GUIDE.md)** - Usage instructions
- **[Developer Guide](DEVELOPMENT.md)** - Development setup
- **[API Reference](API_REFERENCE.md)** - CLI reference

---

**Questions?** Open an issue on [GitHub](https://github.com/OpenCHAMI/legendary-funicular/issues)!
