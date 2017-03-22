#!/usr/bin/env bash

EXIT_CODE_SUCCESS=0
EXIT_CODE_FAIL=1

pushd "install"
./systemd-install.sh || exit ${EXIT_CODE_FAIL}
popd
echo "Done!"
exit ${EXIT_CODE_SUCCESS}