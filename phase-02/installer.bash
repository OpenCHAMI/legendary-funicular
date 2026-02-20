#!/usr/bin/env bash

set -x # remove after prototyping
set -euo pipefail

if [ ! "$UID" == "0" ]; then
    echo "script must be run as the root user"
    exit 1
fi

FILE_VECTOR_CONF=vector-cloud-events.yaml

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
        echo "installing $(basename $path_source)"
        cp -v $path_source $path_target
        cd $path_target
        chown root: $path_source
        chmod $mode_value $path_source
        restorecon -v $path_source
    )
}

fail-if-missing ${FILE_VECTOR_CONF}
mkdir -vp /etc/vector.d /var/lib/vector
install-file ${FILE_VECTOR_CONF} /etc/vector.d 0644

echo "restarting services"
systemctl daemon-reload
systemctl restart openchami-logq-log-writer.service

echo "finished successfully"
