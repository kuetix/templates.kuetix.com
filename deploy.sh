#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd -- "$(dirname -- "$0")" && pwd)"
DEPLOY_HOST=${DEPLOY_HOST:-kuetix.com}
REMOTE_ROOT=${REMOTE_ROOT:-/opt/kuetix/templates}
HEALTH_URL=${HEALTH_URL:-https://templates.kuetix.com/}
VERSION=${VERSION:-$(date -u +%Y%m%d%H%M%S)}

case "$VERSION" in
    ''|*[!A-Za-z0-9._-]*) echo "VERSION contains unsafe characters: $VERSION" >&2; exit 2 ;;
esac

for arg in "$@"; do
    case "$arg" in
        -y|--yes) ;;
        -h|--help)
            echo "Usage: ./deploy.sh [-y|--yes]"
            echo "Publishes templates to ${REMOTE_ROOT}/templates.kuetix.com-${VERSION}."
            exit 0
            ;;
        *) echo "Unknown argument: $arg" >&2; exit 2 ;;
    esac
done

command -v ssh >/dev/null || { echo "ssh is required" >&2; exit 1; }
command -v tar >/dev/null || { echo "tar is required" >&2; exit 1; }

SOURCE_DIR=${SOURCE_DIR:-${ROOT_DIR}/templates}
[ -d "$SOURCE_DIR" ] || { echo "Template directory not found: $SOURCE_DIR" >&2; exit 1; }

RELEASE_DIR="${REMOTE_ROOT}/templates.kuetix.com-${VERSION}"
STAGING_DIR="${REMOTE_ROOT}/.deploy-${VERSION}-$$"

echo "--> Uploading templates release ${VERSION}"
ssh -o ClearAllForwardings=yes "$DEPLOY_HOST" "mkdir -p '${STAGING_DIR}/templates'"
tar -C "$SOURCE_DIR" -czf - . | ssh -o ClearAllForwardings=yes "$DEPLOY_HOST" "tar -xzf - -C '${STAGING_DIR}/templates'"
ssh -o ClearAllForwardings=yes "$DEPLOY_HOST" "
    set -eu
    printf '%s\\n' '${VERSION}' > '${STAGING_DIR}/templates/VERSION'
    if [ -e '${RELEASE_DIR}' ]; then
        echo "Release already exists: ${RELEASE_DIR}" >&2
        exit 1
    fi
    mv '${STAGING_DIR}' '${RELEASE_DIR}'
    tar -czf '${REMOTE_ROOT}/${VERSION}.tar.gz' -C '${RELEASE_DIR}' templates
    ln -sfn '${RELEASE_DIR}/templates' '${REMOTE_ROOT}/${VERSION}.next'
    mv -Tf '${REMOTE_ROOT}/${VERSION}.next' '${REMOTE_ROOT}/${VERSION}'
    ln -sfn '${REMOTE_ROOT}/${VERSION}' '${REMOTE_ROOT}/latest.next'
    mv -Tf '${REMOTE_ROOT}/latest.next' '${REMOTE_ROOT}/latest'
"

if [ -n "$HEALTH_URL" ]; then
    command -v curl >/dev/null || { echo "curl is required for HEALTH_URL" >&2; exit 1; }
    echo "--> Health check: ${HEALTH_URL}"
    curl --fail --silent --show-error --max-time "${HEALTH_TIMEOUT:-30}" "$HEALTH_URL" >/dev/null
fi

echo "--> Templates deployed: ${VERSION}"
