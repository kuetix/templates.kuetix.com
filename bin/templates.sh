#!/usr/bin/env bash

REPO="kuetix/templates.kuetix.com"
ROOT_DIR="/opt/kuetix/templates/"
if [ ! -d "${ROOT_DIR}" ]; then
  echo "Directory ${ROOT_DIR} does not exist, exiting."
  exit 1
fi
cd "${ROOT_DIR}" || exit
TAG_VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | jq -r '.tag_name' | tr -d "v")
if [ "${TAG_VERSION}" == "null" ] || [ "${TAG_VERSION}" == "" ]; then
  TAG_VERSION=$(/usr/bin/ls -1dv v* 2>&1 | grep -v "No such file or directory" | tail -n 1 | sed "s/.tar.gz//" | tr -d "v")
  if [ "${TAG_VERSION}" == "" ]; then
    echo "No version found, exiting."
    exit 1
  fi
fi
TAG="v${TAG_VERSION}"

curl -fL "https://github.com/${REPO}/archive/refs/tags/${TAG}.tar.gz" -o "${TAG}.tar.gz" -# || true
#
ls -la "${TAG}.tar.gz"
echo "Unpack ${TAG}.tar.gz"
rm -fR "${ROOT_DIR}templates.kuetix.com-${TAG_VERSION}"
tar -vxzf "${TAG}.tar.gz"
echo "ln -s \"${ROOT_DIR}templates.kuetix.com-${TAG_VERSION}/templates/\" \"${TAG_VERSION}\""
rm -fR "${TAG_VERSION}"
ln -s "${ROOT_DIR}templates.kuetix.com-${TAG_VERSION}/templates/" "${TAG_VERSION}"
touch "templates.kuetix.com-${TAG_VERSION}/templates/VERSION"
echo "${TAG_VERSION}" > "templates.kuetix.com-${TAG_VERSION}/templates/VERSION"
tar -vczf "${TAG_VERSION}.tar.gz" -C "templates.kuetix.com-${TAG_VERSION}/" "templates"
cp "${TAG_VERSION}.tar.gz" "${TAG_VERSION}/templates.tar.gz"
rm -rf "latest"
ln -s "${ROOT_DIR}${TAG_VERSION}" "latest"
