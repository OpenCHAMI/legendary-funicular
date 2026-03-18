#!/usr/bin/env bash

# SPDX-FileCopyrightText: Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
#
# SPDX-License-Identifier: MIT

# set -x # remove after prototyping
set -euo pipefail

if [ ! "$UID" == "0" ]; then
    echo "script must be run as the root user"
    exit 1
fi

fail-if-missing() {
    if [ ! -f "$1" ]; then
        echo "error: could not locate $1"
        exit 1
    fi
}

install-file() {
    if [ ! "$#" == "3" ]; then
        echo "invalid arg count=$#, expected 3"
        exit 1
    fi
    (
        path_source=$1
        path_target=$2
        mode_value=$3
        basename_source=$(basename $path_source)
        echo "installing $@"
        # cp -v $path_source $path_target
        cp $path_source $path_target
        cd $path_target
        chown root: $basename_source
        chmod $mode_value $basename_source
        restorecon -v $basename_source
    )
}

install-dir() {
    if [ ! "$#" == "3" ]; then
        echo "invalid arg count=$#, expected 3"
        exit 1
    fi
    (
        path_source=$1
        path_target=$2
        mode_value=$3

        if [ ! -d "$path_source" ]; then
            echo "expected \$1=/path/to/directory, received '$path_source'"
            exit 1
        elif [ ! -e "$path_target" ]; then
            mkdir -vp $path_target
        elif [ ! -d "$path_target" ]; then
            echo "expected \$2=/path/to/directory, received '$path_target'"
            exit 1
        fi

        for f in "$path_source"/*; do
            if [ -f "$f" ]; then
                install-file $f $path_target $mode_value
            elif [ -d "$f" ]; then
                install-dir $f "$path_target/$(basename $f)" $mode_value
            else
                echo "unexpected arg with unknown filetype: '$f'"
                exit 1
            fi
        done
    )
}

################################################################################
################################################################################
# BUILD VARS
# Derived from:
# https://github.com/OpenCHAMI/ochami/blob/2247677c4e2a944e230872b5024100078dd8efec/Makefile
################################################################################
################################################################################
GO=/usr/local/go/bin/go

ORG=seantronsen
NAME=openchami-logq/query
IMPORT=github.com/${ORG}/${NAME}/
VERSION=$(git describe --tags --always --dirty --broken --abbrev=0)
TAG=$(git describe --tags --always --abbrev=0)
BRANCH=$(git branch --show-current)
BUILD=$(git rev-parse HEAD)
GOVER=$(${GO} env GOVERSION)
GITSTATE=$(if output=$(git status --porcelain) && [ -n "$output" ]; then echo dirty; else echo clean; fi)
BUILDHOST=$(hostname)
BUILDUSER=$(whoami)

LDFLAGS="-s"
LDFLAGS+=" -X ${IMPORT}internal/version.Version=${VERSION}"
LDFLAGS+=" -X ${IMPORT}internal/version.Tag=${TAG}"
LDFLAGS+=" -X ${IMPORT}internal/version.Branch=${BRANCH}"
LDFLAGS+=" -X ${IMPORT}internal/version.Commit=${BUILD}"
LDFLAGS+=" -X ${IMPORT}internal/version.Date=$(date -Iseconds)"
LDFLAGS+=" -X ${IMPORT}internal/version.GoVersion=${GOVER}"
LDFLAGS+=" -X ${IMPORT}internal/version.GitState=${GITSTATE}"
LDFLAGS+=" -X ${IMPORT}internal/version.BuildHost=${BUILDHOST}"
LDFLAGS+=" -X ${IMPORT}internal/version.BuildUser=${BUILDUSER}"

################################################################################
################################################################################
# INSTALL
################################################################################
################################################################################

EXPECTED=(
    "./deploy/configs/vector.d/transforms/logs_exclusiverouter.yaml"
    "./deploy/configs/vector.d/transforms/logs_sanitizer_json.yaml"
    "./deploy/configs/vector.d/transforms/logs_preparer.yaml"
    "./deploy/configs/vector.d/transforms/logs_sanitizer_logfmt.yaml"
    "./deploy/configs/vector.d/transforms/cloudevents_parser.yaml"
    "./deploy/configs/vector.d/transforms/logs_parser.yaml"
    "./deploy/configs/vector.d/sources/http_cloudevents.yaml"
    "./deploy/configs/vector.d/sources/syslog_udp_logs.yaml"
    "./deploy/configs/vector.d/vector.yaml"
    "./deploy/configs/vector.d/sinks/stdout.yaml"
    "./deploy/configs/vector.d/sinks/s3_logs.yaml"
    "./deploy/configs/vector.d/sinks/s3_events.yaml"
    "./deploy/configs/rsyslog.d/openchami-service-lookup.json"
    "./deploy/configs/rsyslog.d/10-relay-openchami.conf"
    "./deploy/scripts/openchami-logq-versitygw-bootstrap.sh"
    "./deploy/systemd/containers/openchami-logq-log-writer.container"
    "./deploy/systemd/system/openchami-logq-versitygw-bootstrap.service"
    "./deploy/systemd/system/openchami-logq-compaction.timer"
    "./deploy/systemd/system/openchami-logq-compaction.service"
)

echo "verifying existence of expected files"
for f in "${EXPECTED[@]}"; do
    fail-if-missing $f
done

echo "installing"
install-dir deploy/configs /etc 0644
install-dir deploy/systemd/containers /etc/containers/systemd 0644
install-dir deploy/systemd/system /etc/systemd/system 0644
install-dir deploy/scripts /usr/local/libexec 0755
(
    cd compactor
    ${GO} build .
    mv compactor openchami-logq-compactor
    install-file ./openchami-logq-compactor /usr/local/libexec 0755
)
(
    cd query
    ${GO} build -v -o openchami-logq -ldflags="${LDFLAGS}"
    install-file ./openchami-logq /usr/local/bin 0755
)

echo "creating service work directories"
mkdir -vp /var/lib/vector

echo "restarting dependent services"
systemctl daemon-reload
systemctl restart openchami-logq-versitygw-bootstrap.service
systemctl restart rsyslog.service
systemctl restart openchami-logq-log-writer.service
systemctl restart openchami-logq-compaction.timer

echo "finished successfully"
