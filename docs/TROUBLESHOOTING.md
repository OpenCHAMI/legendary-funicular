<!--
SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC

SPDX-License-Identifier: MIT
-->

# Troubleshooting Guide

**Project:** openchami-logq
**Version:** 1.0
**Last Updated:** June 10, 2026

---

## Common Problems and Solutions

**Query fails with "S3 access denied":**

```bash
# Check credentials
aws s3 ls s3://openchami-logs-daily --endpoint-url=$S3_ENDPOINT

# Verify environment
openchami-logq-query inspect config
```

**Compactor fails:**

```bash
# Run with verbose logging
openchami-logq-compactor --verbose

# Check S3 permissions
# Compactor needs: s3:GetObject, s3:PutObject, s3:DeleteObject
```

**Query returns no results:**

```bash
# Check available dates
openchami-logq-query inspect dates --stream logs

# Verify data exists
aws s3 ls s3://openchami-logs-daily/logs/ --endpoint-url=$S3_ENDPOINT
```

For additional troubleshooting, check the project's GitHub issues or open
a new issue with details about your environment and error messages.
