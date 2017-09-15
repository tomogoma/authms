#!/usr/bin/env bash
NAME="authms"
VERSION="v1"
DESCRIPTION="Authentication Micro-Service"
CANONICAL_NAME="${NAME}${VERSION}"
CONF_DIR="/etc/${NAME}"
CONF_FILE="${CONF_DIR}/${CANONICAL_NAME}.conf.yml"
INSTALL_DIR="/usr/local/bin"
INSTALL_FILE="${INSTALL_DIR}/${CANONICAL_NAME}"
UNIT_NAME="${CANONICAL_NAME}.service"
UNIT_FILE="/etc/systemd/system/${UNIT_NAME}"

printDetails() {
    echo "name:              $NAME"
    echo "version:           $VERSION"
    echo "description:       $DESCRIPTION"
    echo "canonical name:    $CANONICAL_NAME"
    echo "conf dir:          $CONF_DIR"
    echo "conf file:         $CONF_FILE"
    echo "install file:      $INSTALL_FILE"
    echo "systemd unit file: $UNIT_FILE"
}