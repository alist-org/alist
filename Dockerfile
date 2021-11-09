FROM alpine as builder
LABEL stage=go-builder
WORKDIR /app/
COPY ./ ./
RUN apk add --no-cache bash git go=1.17.3-r0 gcc musl-dev; \
    sh build.sh docker

FROM alpine
LABEL MAINTAINER="i@nn.ci"
WORKDIR /opt/alist/
COPY --from=builder /app/bin/alist ./
EXPOSE 5244
CMD [ "./alist" ]