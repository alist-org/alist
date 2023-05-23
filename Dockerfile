##############################################################################################################################
##                                                                                                                          ##
##  It is recommended to use buildx for building:                                                                           ##
##                                                                                                                          ##
##  // Note: you can use the standard `docker build` command, but there is no multi architecture support                    ##
##                                                                                                                          ##
##  // Create buildx instance                                                                                               ##
##  docker buildx create --driver docker-container --name builder --bootstrap --use                                         ##
##                                                                                                                          ##
##  // Login to Rregistry                                                                                                   ##
##  docker login [REGSITRY_ADDRESS]                                                                                         ##
##                                                                                                                          ##
##  // Build the docker image (both x86 and arm64 are supported)                                                            ##
##  docker buildx build --platform=linux/amd64,linux/arm64 --push -t [REGSITRY_ADDRESS]/REGSITRY_USERNAME/alist .           ##
##                                                                                                                          ##
##############################################################################################################################

# Build the Alist executable
#
# Note: At this stage, only the compiled executable file is kept, 
# and the metadata of the image will be set in the next stage.
FROM golang:1.20-alpine3.18 AS builder

ENV WEBDIST_URL https://github.com/alist-org/alist-web/releases/latest/download/dist.tar.gz
ENV WEBDIST_API_URL https://api.github.com/repos/alist-org/alist-web/releases/latest

COPY ./ /go/src/alist

WORKDIR /go/src/alist

RUN set -ex \
    && apk add --no-cache bash curl git jq gcc musl-dev \
    && export builtAt="$(date +'%F %T %z')" \
    && export goVersion="$(go version | sed 's/go version //')" \
    && export gitAuthor="Xhofe <i@nn.ci>" \
    && export gitCommit=$(git rev-parse HEAD) \
    && export version=$(git describe --tags --always --dirty) \
    && export webVersion=$(curl -sSL "${WEBDIST_API_URL}" | jq -r '.tag_name') \
    && curl -fsSL ${WEBDIST_URL} > /tmp/dist.tar.gz \
    && rm -rf public/dist && tar -C public/ -xvf /tmp/dist.tar.gz \
    && go install -trimpath -tags=jsoniter -ldflags="-w -s \
            -X 'github.com/alist-org/alist/v3/internal/conf.BuiltAt=${builtAt}' \
            -X 'github.com/alist-org/alist/v3/internal/conf.GoVersion=${goVersion}' \
            -X 'github.com/alist-org/alist/v3/internal/conf.GitAuthor=${gitAuthor}' \
            -X 'github.com/alist-org/alist/v3/internal/conf.GitCommit=${gitCommit}' \
            -X 'github.com/alist-org/alist/v3/internal/conf.Version=${version}' \
            -X 'github.com/alist-org/alist/v3/internal/conf.WebVersion=${webVersion}'"


# Build the final Docker image.
FROM alpine:3.18 AS dist

LABEL maintainer="i@nn.ci"
LABEL repository="https://github.com/alist-org/alist"
LABEL description="A file list program that supports multiple storage, powered by Gin and Solidjs."
LABEL licenses="AGPL-3.0"

LABEL org.opencontainers.image.authors="i@nn.ci"
LABEL org.opencontainers.image.source="https://github.com/alist-org/alist"
LABEL org.opencontainers.image.description="A file list program that supports multiple storage, powered by Gin and Solidjs."
LABEL org.opencontainers.image.licenses="AGPL-3.0"

# Multiple ENV instructions do not take up extra space;
# instead, they can increase readability.
ENV PUID=0
ENV PGID=0
ENV UMASK=022

RUN apk add --no-cache bash ca-certificates tzdata su-exec

COPY --from=builder /go/bin/alist /usr/bin
COPY entrypoint.sh /entrypoint.sh

WORKDIR /opt/alist/

VOLUME /opt/alist/data/

EXPOSE 5244

HEALTHCHECK CMD echo -e "GET / HTTP/1.1\r\n" | nc -w 5 localhost 5244

CMD ["/entrypoint.sh"]
