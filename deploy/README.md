<!--
SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC

SPDX-License-Identifier: MIT
-->

# Deployment Examples

Example configuration files and deployment templates for openchami-logq.

## Directory Structure

```
deploy/
├── docker/                    # Docker Compose examples
│   └── docker-compose.yml     # Complete stack (VersityGW, Vector, Compactor)
├── systemd/                   # Systemd unit files
│   ├── containers/            # Podman Quadlet container units
│   ├── system/                # Traditional systemd units
│   ├── openchami-logq-compactor.container  # Quadlet compactor
│   ├── openchami-logq-compactor.timer      # Compaction schedule
│   └── vector.service         # Vector log collector
├── vector/                    # Vector configuration examples
│   └── vector.yaml            # Vector → S3 pipeline
└── scripts/                   # Helper scripts
    └── openchami-logq-versitygw-bootstrap.sh  # VersityGW setup
```

## Usage

### Docker Compose

Complete stack with VersityGW, Vector, and Compactor:

```bash
cd deploy/docker
docker-compose up -d
```

See [Operations Guide](../docs/OPERATIONS.md#option-2-docker-compose) for details.

### Podman Quadlet

Production deployment with systemd integration:

```bash
# Copy Quadlet files
sudo cp deploy/systemd/openchami-logq-compactor.container \
  /etc/containers/systemd/

sudo cp deploy/systemd/openchami-logq-compactor.timer \
  /etc/containers/systemd/

# Enable and start
systemctl daemon-reload
systemctl enable --now openchami-logq-compactor.timer
```

See [Operations Guide](../docs/OPERATIONS.md#option-1-podman-quadlet) for details.

### Vector

Copy example configuration:

```bash
sudo cp deploy/vector/vector.yaml /etc/vector/vector.yaml
sudo cp deploy/systemd/vector.service /etc/systemd/system/
systemctl daemon-reload
systemctl enable --now vector
```

## Customization

All files contain placeholder values that must be updated:

- **S3 credentials**: Replace `log-writer-key`, `log-compactor-key`, etc.
- **S3 endpoint**: Replace `http://localhost:7070` with your S3 URL
- **Bucket names**: Replace `openchami-logs-raw` and `openchami-logs-daily`

## Security Notes

- Store credentials in environment files with `chmod 600`
- Use separate IAM users for each component
- Enable SSL/TLS in production (`S3_SSL=true`)
- Rotate credentials regularly

## Documentation

- **[Operations Guide](../docs/OPERATIONS.md)** - Complete deployment instructions
- **[Architecture](../docs/ARCHITECTURE.md)** - System design and component overview
- **[Main README](../README.md)** - Quick start and configuration reference
