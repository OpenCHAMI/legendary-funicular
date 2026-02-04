#!/usr/bin/env bash

set -x # remove after prototyping
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

fail-if-missing openchami-logq-versitygw-bootstrap.service
fail-if-missing openchami-logq-versitygw-bootstrap.sh

echo "installing bootstrap script"
(
	target=openchami-logq-versitygw-bootstrap.sh
	destination=/usr/local/libexec/
	cp -v $target $destination
	cd $destination
	chown root: $target
	chmod 0755 $target
	restorecon -v $target
)

echo "installing bootstrap service"
(
	target=openchami-logq-versitygw-bootstrap.service
	destination=/etc/systemd/system/
	cp -v $target $destination
	cd $destination
	chown root: $target
	chmod 0644 $target
	restorecon -v $target
)

echo "finished successfully"
