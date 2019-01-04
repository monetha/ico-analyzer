#!/bin/bash


set -e -o pipefail

if [ -z "$CI_PIPELINE_ID" ]
  then
    echo "No CI_PIPELINE_ID supplied."
    exit 1
fi

GOLANG_VERSION=1.11
PACKAGE_NAME=github.com/monetha/ico-analyzer
PACKAGE_FULL_PATH=/go/src/$PACKAGE_NAME
VERSION=$(date +"%Y.%m.%d").$CI_PIPELINE_ID

docker run -i --rm \
-v "$PWD":$PACKAGE_FULL_PATH \
-w $PACKAGE_FULL_PATH \
golang:$GOLANG_VERSION /bin/bash << COMMANDS
set -e -o pipefail
[[ $EUID -ne 0 ]] && echo Creating host user $(id -un) in container...
[[ $EUID -ne 0 ]] && addgroup --gid $(id -g) $(id -gn)
[[ $EUID -ne 0 ]] && adduser --disabled-password --gecos "" --no-create-home --home $PACKAGE_FULL_PATH --uid $(id -u) --gid $(id -g) $(id -un)
[[ $EUID -ne 0 ]] && adduser $(id -un) sudo
[[ $EUID -ne 0 ]] && adduser $(id -un) root
[[ $EUID -ne 0 ]] && echo "ENV_SUPATH   PATH=\$PATH" >> /etc/login.defs
[[ $EUID -ne 0 ]] && echo "ENV_PATH     PATH=\$PATH" >> /etc/login.defs
[[ $EUID -ne 0 ]] && echo Switching to user $(id -un)...
[[ $EUID -ne 0 ]] && su -m $(id -un)
apt-get update
apt-get install -y libxml2-dev
set -e -o pipefail
HOME=$PACKAGE_FULL_PATH
cd $PACKAGE_FULL_PATH
make -j4 VERSION="$VERSION"
COMMANDS

echo Done.
