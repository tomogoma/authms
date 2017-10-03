#!/usr/bin/env bash

source vars.sh

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
    echo "templates dir:     $TPL_DIR"
}

mkdir -p "${CONF_DIR}" || exit 1
if [ ! -f "${CONF_FILE}" ]; then
    cp "conf.yml" "${CONF_FILE}" || exit 1
fi

mkdir -p "${TPL_DIR}" || exit 1
if [ ! -f "${EMAIL_INVITE_TPL}" ]; then
    cp "invitation_email.html" "${EMAIL_INVITE_TPL}" || exit 1
fi
if [ ! -f "${PHONE_INVITE_TPL}" ]; then
    cp "invitation_sms.tpl" "${PHONE_INVITE_TPL}" || exit 1
fi
if [ ! -f "${EMAIL_RESET_PASS_TPL}" ]; then
    cp "reset_pass_email.html" "${EMAIL_RESET_PASS_TPL}" || exit 1
fi
if [ ! -f "${PHONE_RESET_PASS_TPL}" ]; then
    cp "reset_pass_sms.tpl" "${PHONE_RESET_PASS_TPL}" || exit 1
fi
if [ ! -f "${EMAIL_VERIFY_TPL}" ]; then
    cp "verify_email.html" "${EMAIL_VERIFY_TPL}" || exit 1
fi
if [ ! -f "${PHONE_VERIFY_TPL}" ]; then
    cp "verify_sms.tpl" "${PHONE_VERIFY_TPL}" || exit 1
fi

mkdir -p "${INSTALL_DIR}" || exit 1
cp -f ../bin/app "${INSTALL_FILE}" || exit 1
printUnitFile
systemctl enable "${UNIT_NAME}"
systemctl daemon-reload

printDetails
