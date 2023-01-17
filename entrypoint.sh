#!/bin/bash

chown -R ${PUID}:${PGID} /opt/alist/

umask ${UMASK}

exec su-exec ${PUID}:${PGID} nohup aria2c --enable-rpc --rpc-allow-origin-all > /dev/null 2>&1 &

exec su-exec ${PUID}:${PGID} ./alist server --no-prefix