#!/bin/bash

chown -R ${PUID}:${PGID} /opt/alist/

umask ${UMASK}
# 指定 Python 程序的工作目录
PYTHON_DIR="/app/auto_pikpak"

# 启动 Python 程序
(
  cd "$PYTHON_DIR" && /app/auto_pikpak/venv/bin/python3 server.py &
)

if [ "$1" = "version" ]; then
  ./alist version
else
  exec ./alist server --no-prefix
fi