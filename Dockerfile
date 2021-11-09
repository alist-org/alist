FROM golang:alpine as builder
LABEL stage=go-builder
WORKDIR /app/
COPY ./ ./
RUN bash build.sh docker

FROM alpine
LABEL MAINTAINER="i@nn.ci"
WORKDIR /opt/alist/
COPY --from=builder /alist/bin/alist ./
EXPOSE 5244
CMD [ "./alist" ]