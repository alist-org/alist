#!/usr/bin/env bash

set -e

chown -R ${PUID}:${PGID} /opt/alist/
umask ${UMASK}
exec su-exec ${PUID}:${PGID} alist server --no-prefix
