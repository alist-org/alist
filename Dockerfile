FROM ubuntu AS build_alist
WORKDIR /app/
# RUN apt update && apt install bash curl gcc git go musl-dev
# 安装基本工具
RUN apt update && \
    apt install -y software-properties-common && \
    add-apt-repository ppa:longsleep/golang-backports && \
    apt update && \
    apt install -y bash curl gcc git golang-go musl-dev
COPY go.mod go.sum ./
RUN go mod download
COPY ./ ./
RUN bash build.sh release docker

FROM ubuntu AS install_py
RUN apt update && \
    apt install -y curl python3 python3-pip python3.12-venv
WORKDIR /app/auto_pikpak/
# RUN curl -O -L https://raw.githubusercontent.com/wangjunkai2022/auto_pikpak/main/requirements.txt requirements.txt
# RUN curl -O -L https://raw.githubusercontent.com/wangjunkai2022/auto_pikpak/main/requirements.txt requirements.txt
# 创建虚拟环境并激活
RUN python3 -m venv venv && \
    . venv/bin/activate && \
    pip install \
    PyYAML \
    selenium \
    pyTelegramBotAPI \
    pyrclone \
    httpx \
    numpy \
    opencv_python \
    opencv_python_headless \
    ultralytics \
    2captcha-python \
    Flask

FROM ubuntu as pikpak_server
RUN apt update && \
    apt install -y git
WORKDIR /app
RUN git clone https://github.com/wangjunkai2022/auto_pikpak.git --depth 1
WORKDIR /app/auto_pikpak
COPY --from=install_py /app/auto_pikpak/venv /app/auto_pikpak/venv

FROM ubuntu
ARG INSTALL_FFMPEG=false
LABEL MAINTAINER="i@nn.ci"

WORKDIR /opt/alist/

RUN apt update && \
    apt upgrade -y && \
    apt install -y bash ca-certificates tzdata ffmpeg

# 复制 auto_pikpak 到第二阶段
COPY --from=pikpak_server /app/auto_pikpak /app/auto_pikpak
COPY --from=build_alist /app/bin/alist ./
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh && /entrypoint.sh version

ENV PUID=0 PGID=0 UMASK=022
VOLUME /opt/alist/data/
EXPOSE 5244 5245
CMD [ "/entrypoint.sh" ]