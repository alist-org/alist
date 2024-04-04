FROM alpine:edge as builder
LABEL stage=go-builder
WORKDIR /app/
RUN apk add --no-cache bash curl gcc git go musl-dev
COPY go.mod go.sum ./
RUN go mod download
COPY ./ ./
RUN bash build.sh release docker

FROM alpine:3.19
LABEL MAINTAINER="i@nn.ci"
VOLUME /opt/alist/data/
WORKDIR /opt/alist/
COPY --from=builder /app/bin/alist ./
COPY entrypoint.sh /entrypoint.sh
RUN sed -i "s/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g" /etc/apk/repositories && \
    apk upgrade --no-cache && \
    apk add --no-cache bash ca-certificates su-exec tzdata alpine-conf && \
    /sbin/setup-timezone -z Asia/Shanghai && \
    chmod +x /entrypoint.sh && \
    apk del alpine-conf && rm -rf /var/cache/apk/*
ENV PUID=0 PGID=0 UMASK=022
EXPOSE 5244 5245
CMD [ "/entrypoint.sh" ]