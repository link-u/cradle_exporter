#! /bin/bash -eux

set -eux

SCRIPT_PATH=$(cd $(dirname $(readlink -f $0)) && pwd)
cd "${SCRIPT_PATH}"
cd ..

fakeroot debian/rules clean
fakeroot debian/rules binary

set +eu
cp -f ../*.deb ../*.ddeb . || true
