#!/bin/bash

chown -R ${PUID}:${PGID} /opt/alist/

umask ${UMASK}

if [ "$1" = "version" ]; then
  ./alist version
else
  exec su-exec ${PUID}:${PGID} ./alist server --no-prefix
fi