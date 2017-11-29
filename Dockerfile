FROM golang:1.9-alpine as builder
COPY . /go/src/github.com/hectorqin/local-mapping
WORKDIR /go/src/github.com/hectorqin/local-mapping
RUN set -ex \
    && go install --ldflags '-extldflags "-static"'
CMD ["/go/bin/local-mapping"]

FROM alpine
RUN apk update \
    && mkdir -p /run/docker/plugins
COPY --from=builder /go/bin/local-mapping .
CMD ["local-mapping"]