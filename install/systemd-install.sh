#!/usr/bin/env bash

./uninstall.sh || exit 1

NAME="authms"
BUILD_NAME="${NAME}-installer-version"
CONF_DIR="/etc/${NAME}"
CONF_FILE="${CONF_DIR}/${NAME}.conf.yml"
INSTALL_DIR="/usr/local/bin"
UNIT_FILE="/etc/systemd/system/${NAME}.service"

echo "Begin install"

mkdir -p "${CONF_DIR}" || exit 1
cp "${NAME}.conf.yml" "${CONF_FILE}" || exit 1

mkdir -p "${INSTALL_DIR}" || exit 1
cp -f "${BUILD_NAME}" "${INSTALL_DIR}/${NAME}" || exit 1

cp -f "${NAME}.service" "${UNIT_FILE}" || exit 1
systemctl enable "${NAME}.service"

echo "Config file is at '${CONF_FILE}"
echo "Install complete"