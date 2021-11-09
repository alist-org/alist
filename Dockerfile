FROM golang:alpine as builder
LABEL stage=go-builder
WORKDIR /app/
COPY ./ ./
RUN sh build.sh docker

FROM alpine
LABEL MAINTAINER="i@nn.ci"
WORKDIR /opt/alist/
COPY --from=builder /app/bin/alist ./
EXPOSE 5244
CMD [ "./alist" ]