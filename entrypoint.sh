#!/bin/bash

chown -R ${PUID}:${PGID} /opt/alist/

umask ${UMASK}

exec su-exec ${PUID}:${PGID} ./alist server --no-prefix