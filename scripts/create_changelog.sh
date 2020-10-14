#! /bin/bash

set -eu

SCRIPT_PATH="$(cd "$(dirname "$(readlink -f "$0")")" && pwd)"
cd "${SCRIPT_PATH}"
cd ..

codename=$(lsb_release -cs)

cp "scripts/changelog_template" "debian/changelog"
# replace date
sed -i -r "s/%DATE%/$(LC_ALL=C TZ=JST-9 date '+%a, %d %b %Y %H:%M:%S %z')/g" debian/changelog

# set debian package version
version=$(git describe --tags || (echo -n "v0.0.0-" && git describe --always))
dpkg_version=${version:1}.$(TZ=JST-9 date +%Y%m%d)+${codename}
sed -i -r "s/%VERSION%/${dpkg_version}/g" debian/changelog
