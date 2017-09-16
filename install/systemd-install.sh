#!/usr/bin/env bash

source vars.sh
./systemd-uninstall.sh || exit 1

printUnitFile() {
read -r -d '' UNIT_CONTENTS  << EOF
[Unit]
Description=${DESCRIPTION}
Wants=network-online.target
After=network.target network-online.target

[Install]
WantedBy=multi-user.target

[Service]
ExecStart=${INSTALL_FILE}
SyslogIdentifier=${CANONICAL_NAME}
Restart=always
RestartSec=10
EOF
echo "$UNIT_CONTENTS" > ${UNIT_FILE} || exit 1
}

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

mkdir -p "${CONF_DIR}" || exit 1
if [ ! -f "${CONF_FILE}" ]; then
    cp "conf.yml" "${CONF_FILE}" || exit 1
fi
mkdir -p "${INSTALL_DIR}" || exit 1
cp -f ../bin/app "${INSTALL_FILE}" || exit 1
printUnitFile
systemctl enable "${UNIT_NAME}"
systemctl daemon-reload

printDetails
