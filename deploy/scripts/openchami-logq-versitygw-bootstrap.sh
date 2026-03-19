#!/usr/bin/env bash

# SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
#
# SPDX-License-Identifier: MIT

set -euo pipefail

# ------------------------------------------------------------------------------
# Minimal bootstrap to provision users and buckets for openchami-logq
#
# Creates users:
#   - log-writer
#   - log-compactor
#   - log-reader
#
# For each user:
#   - Generate random access/secret keys
#     (stored in /etc/versitygw/users.d/<user>.env)
#   - Create IAM user with role "user"
#
# Creates buckets:
#   - openchami-logs-raw
#   - openchami-logs-daily
#
# openchami-logs-raw:
#   - Private (authorized users only)
#   - log-writer: write-only
#   - log-compactor: read
#
# openchami-logs-daily:
#   - Private (authorized users only)
#   - log-compactor: write-only
#   - log-reader: read
#
# All operations are idempotent and safe to re-run.
# ------------------------------------------------------------------------------
################################################################################
################################################################################
# SCHEMA TABLE: ACCOUNT PURPOSE | USE CASES
################################################################################
################################################################################
# | Identity        | `openchami-logs-raw`       | `openchami-logs-daily`         | Notes                                    |
# | --------------- | -------------------------- | ------------------------------ | ---------------------------------------- |
# | `log-writer`    | **write-only** (PutObject) | none                           | Vector / Fluent Bit pushes raw           |
# | `log-compactor` | **read** (Get/List)        | **write** (Put + maybe Delete) | Batch job reads raw → writes Parquet     |
# | `log-reader`    | **read-only** (Get/List)   | **read-only** (Get/List)       | DuckDB CLI / query utility reads Parquet |
################################################################################
################################################################################

GATEWAY_ENDPOINT="${GATEWAY_ENDPOINT:-http://127.0.0.1:7070}"
CONTAINER_NAME="${CONTAINER_NAME:-versitygw}"

ROOT_ACCESS="${ROOT_ACCESS_KEY:?ROOT_ACCESS_KEY not set}"
ROOT_SECRET="${ROOT_SECRET_KEY:?ROOT_SECRET_KEY not set}"

USERS_DIR=/etc/versitygw/users.d

# ------------------------------
# Define your users here
# ------------------------------
BUCKETS=(
    "openchami-logs-raw"
    "openchami-logs-daily"
)
USERS=(
    "log-writer"
    "log-compactor"
    "log-reader"
)
BUCKET_RAW="${BUCKETS[0]}"
BUCKET_DAILY="${BUCKETS[1]}"
USER_LOG_WRITER="${USERS[0]}"
USER_LOG_COMPACTOR="${USERS[1]}"
USER_LOG_READER="${USERS[2]}"

# ------------------------------------------------------------------------------
# Wait for gateway to be up
# ------------------------------------------------------------------------------
echo "bootstrap: waiting for VersityGW at ${GATEWAY_ENDPOINT}..."
for i in {1..60}; do
    if curl -sSf "${GATEWAY_ENDPOINT}" >/dev/null 2>&1; then
        echo "bootstrap: gateway is up."
        break
    fi
    sleep 1
done

# ------------------------------------------------------------------------------
# Helper to call versitygw admin inside container
# ------------------------------------------------------------------------------
vgw_admin() {
    podman exec "${CONTAINER_NAME}" \
        versitygw admin \
        --access "${ROOT_ACCESS}" \
        --secret "${ROOT_SECRET}" \
        --endpoint-url "${GATEWAY_ENDPOINT}" \
        "$@"
}

# ------------------------------------------------------------------------------
# Helper to get user access string from the corresponding secrets file
# ------------------------------------------------------------------------------
get_user_access() {
    grep '^VGW_ACCESS_KEY=' "${USERS_DIR}/${1}.env" | cut -d= -f2
}

# todo: this is already done by the `gen-secrets.service` we depend on
# so we should probably fail non-zero here as it would violate that contract
# # Ensure dir exists
# mkdir -p "${USERS_DIR}"
# chmod 700 "${USERS_DIR}"
# chown root:root "${USERS_DIR}"

# todo: ibid
# # ------------------------------------------------------------------------------
# # Configure root AWS profile for bucket operations
# # ------------------------------------------------------------------------------
ROOT_PROFILE="vgw-root"
#
# mkdir -p /root/.aws
# chmod 700 /root/.aws
#
# cat >/root/.aws/credentials <<EOF
# [${ROOT_PROFILE}]
# aws_access_key_id     = ${ROOT_ACCESS}
# aws_secret_access_key = ${ROOT_SECRET}
# EOF
#
# chmod 600 /root/.aws/credentials

# ------------------------------------------------------------------------------
# Main loop: per-user IAM
# ------------------------------------------------------------------------------
for user in "${USERS[@]}"; do
    echo "bootstrap: processing '${user}'"

    user_file="${USERS_DIR}/${user}.env"

    # 1. Generate access/secret if not already created
    if [[ ! -f "${user_file}" ]]; then
        echo "  generating new credentials"
        umask 077
        access=$(openssl rand -hex 16)
        secret=$(openssl rand -hex 32)
        cat >"${user_file}" <<EOF
VGW_USER=${user}
VGW_ACCESS_KEY=${access}
VGW_SECRET_KEY=${secret}

S3_ACCESS_KEY=${access}
S3_SECRET_KEY=${secret}
EOF
        chmod 600 "${user_file}"
        chown root:root "${user_file}"
    else
        # shellcheck disable=SC1090
        . "${user_file}"
        access="${VGW_ACCESS_KEY}"
        secret="${VGW_SECRET_KEY}"
        echo "  using existing credentials"
    fi

    # 2. Ensure IAM user exists (role=default 'user')
    if vgw_admin list-users 2>/dev/null | awk 'NR>2 {print $1}' | grep -qx "${access}"; then
        echo "  IAM user exists"
    else
        echo "  creating IAM user"
        vgw_admin create-user \
            --access "${access}" \
            --secret "${secret}" \
            --role user
    fi

    echo "  done for ${user}"
done

# todo: should we hash the bucket names to make them less obvious?
for bucket in "${BUCKETS[@]}"; do
    if aws --profile "${ROOT_PROFILE}" \
        --endpoint-url "${GATEWAY_ENDPOINT}" \
        s3api head-bucket --bucket "${bucket}" >/dev/null 2>&1; then
        echo "  bucket exists (${bucket})"
    else
        echo "  creating bucket ${bucket}"
        aws --profile "${ROOT_PROFILE}" \
            --endpoint-url "${GATEWAY_ENDPOINT}" \
            s3api create-bucket --bucket "${bucket}"
    fi

    # 4. idempotency: ensure expected bucket owner
    # with versitygw, this also resets all ACL and policy configurations to the
    # default (i.e., private state with contents accessible only to the bucket
    # owner). in other words, all previous ACL and access policies get nuked.
    echo "  assigning bucket owner"
    vgw_admin change-bucket-owner \
        --bucket "${bucket}" \
        --owner "${ROOT_ACCESS}"

    # 5. idempotency: ensure bucket owner type => prefer owner
    # with versitygw (and modern aws), this is required to enable use of both
    # ACLs and access policies
    echo "  configuring bucket ownership controls"
    aws --profile "${ROOT_PROFILE}" \
        --endpoint-url "${GATEWAY_ENDPOINT}" \
        s3api put-bucket-ownership-controls --bucket "${bucket}" \
        --ownership-controls '{ "Rules": [ { "ObjectOwnership": "BucketOwnerPreferred" } ] }'
done

# 6. configure expected bucket level ACLs
# todo: the syntax we use for this is always the same
# => create macro and replace for source brevity
# RAW
aws --profile "${ROOT_PROFILE}" \
    --endpoint-url "${GATEWAY_ENDPOINT}" \
    s3api put-bucket-acl \
    --bucket "${BUCKET_RAW}" \
    --grant-write "$(get_user_access ${USER_LOG_WRITER})" \
    --grant-write "$(get_user_access ${USER_LOG_COMPACTOR})" \
    --grant-read "$(get_user_access ${USER_LOG_WRITER})" \
    --grant-read "$(get_user_access ${USER_LOG_COMPACTOR})" \
    --grant-read "$(get_user_access ${USER_LOG_READER})"

# DAILY
aws --profile "${ROOT_PROFILE}" \
    --endpoint-url "${GATEWAY_ENDPOINT}" \
    s3api put-bucket-acl \
    --bucket "${BUCKET_DAILY}" \
    --grant-write "$(get_user_access ${USER_LOG_COMPACTOR})" \
    --grant-read "$(get_user_access ${USER_LOG_READER})"

# 7. configure expected bucket level Policies

# use ol' reliable temp workdir strategy
PATH_WORK=$(mktemp -d)
(
    # configure lazy cleanup on subshell exit
    trap 'rm -rf "${PATH_WORK}"' EXIT
    cd "${PATH_WORK}"
    TEMP_JSON_POLICY="policy.json"

    # RAW
    cat <<EOF >"${TEMP_JSON_POLICY}"
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": "$(get_user_access ${USER_LOG_WRITER})",
      "Action": ["s3:PutObject"],
      "Resource": ["arn:aws:s3:::${BUCKET_RAW}/*"]
    },
    {
      "Effect": "Allow",
      "Principal": "$(get_user_access ${USER_LOG_WRITER})",
      "Action": ["s3:ListBucket"],
      "Resource": ["arn:aws:s3:::${BUCKET_RAW}"]
    },
    {
      "Effect": "Allow",
      "Principal": "$(get_user_access ${USER_LOG_COMPACTOR})",
      "Action": ["s3:GetObject"],
      "Resource": ["arn:aws:s3:::${BUCKET_RAW}/*"]
    },
    {
      "Effect": "Allow",
      "Principal": "$(get_user_access ${USER_LOG_COMPACTOR})",
      "Action": ["s3:DeleteObject"],
      "Resource": ["arn:aws:s3:::${BUCKET_RAW}/*"]
    },
    {
      "Effect": "Allow",
      "Principal": "$(get_user_access ${USER_LOG_COMPACTOR})",
      "Action": ["s3:ListBucket"],
      "Resource": ["arn:aws:s3:::${BUCKET_RAW}"]
    },
    {
      "Effect": "Allow",
      "Principal": "$(get_user_access ${USER_LOG_READER})",
      "Action": ["s3:GetObject"],
      "Resource": ["arn:aws:s3:::${BUCKET_RAW}/*"]
    },
    {
      "Effect": "Allow",
      "Principal": "$(get_user_access ${USER_LOG_READER})",
      "Action": ["s3:ListBucket"],
      "Resource": ["arn:aws:s3:::${BUCKET_RAW}"]
    }
  ]
}
EOF
    aws --profile "${ROOT_PROFILE}" \
        --endpoint-url "${GATEWAY_ENDPOINT}" \
        s3api put-bucket-policy \
        --bucket "${BUCKET_RAW}" \
        --policy file://"${TEMP_JSON_POLICY}"

    # RAW
    cat <<EOF >"${TEMP_JSON_POLICY}"
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": "$(get_user_access ${USER_LOG_COMPACTOR})",
      "Action": ["s3:PutObject"],
      "Resource": ["arn:aws:s3:::${BUCKET_DAILY}/*"]
    },
    {
      "Effect": "Allow",
      "Principal": "$(get_user_access ${USER_LOG_READER})",
      "Action": ["s3:GetObject"],
      "Resource": ["arn:aws:s3:::${BUCKET_DAILY}/*"]
    },
    {
      "Effect": "Allow",
      "Principal": "$(get_user_access ${USER_LOG_READER})",
      "Action": ["s3:ListBucket"],
      "Resource": ["arn:aws:s3:::${BUCKET_DAILY}"]
    }
  ]
}
EOF
    aws --profile "${ROOT_PROFILE}" \
        --endpoint-url "${GATEWAY_ENDPOINT}" \
        s3api put-bucket-policy \
        --bucket "${BUCKET_DAILY}" \
        --policy file://"${TEMP_JSON_POLICY}"
)

echo "bootstrap: COMPLETE"
