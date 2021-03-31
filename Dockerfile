FROM alpine:latest
LABEL MAINTAINER="zhangyi@murphyyi.com"

WORKDIR /app
COPY ./alist .

EXPOSE 5244

# ENTRYPOINT ./alist
ENTRYPOINT  ["./alist"]