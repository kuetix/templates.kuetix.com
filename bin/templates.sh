#!/usr/bin/env bash

cd /opt/kuetix/templates/ || exit
TAG_VERSION=$(/usr/bin/ls -1dv v* 2>&1 | grep -v "No such file or directory" | tail -n 1 | sed "s/.tar.gz//" | tr -d "v")
if [ "${TAG_VERSION}" == "" ]; then
  echo "No version found, exiting."
  exit 1
fi
TAG="v${TAG_VERSION}"
REPO="kuetix/templates.kuetix.com"

curl -fL "https://github.com/${REPO}/archive/refs/tags/${TAG}.tar.gz" -o "${TAG}.tar.gz" -# || true
#
ls -la "${TAG}.tar.gz"
echo "Unpack ${TAG}.tar.gz"
tar -vxzf "${TAG}.tar.gz"
ln -s "templates.kuetix.com-${TAG_VERSION}/templates" "${TAG_VERSION}"
tar -vczf "${TAG_VERSION}.tar.gz" -C "templates.kuetix.com-${TAG_VERSION}/" "templates"
cp "${TAG_VERSION}.tar.gz" "${TAG_VERSION}/templates.tar.gz"
