FROM alpine:edge as builder
LABEL stage=go-builder
WORKDIR /app/
COPY ./ ./
RUN apk add --no-cache bash git go gcc musl-dev curl; \
    bash build.sh release docker

FROM alpine:edge
LABEL MAINTAINER="i@nn.ci"
VOLUME /opt/alist/data/
WORKDIR /opt/alist/
COPY --from=builder /app/bin/alist ./
COPY entrypoint.sh /entrypoint.sh
RUN apk add ca-certificates bash su-exec
ENV PUID=1000 PGID=1000 UMASK=022
EXPOSE 5244
ENTRYPOINT [ "/entrypoint.sh" ]