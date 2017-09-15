#!/usr/bin/env bash

source vars.sh

if [ -f "$UNIT_FILE" ]; then
	systemctl stop  "${UNIT_NAME}" >/dev/null
	rm -f "${UNIT_FILE}" || exit 1
    systemctl daemon-reload
fi
if [ -f "$INSTALL_FILE" ]; then
	rm -f "${INSTALL_FILE}" || exit 1
fi
if [ -f "$CONF_FILE" ]; then
    echo "config file at '${CONF_FILE}' left intact intentionally"
fi
