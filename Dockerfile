FROM golang:1-alpine AS base

WORKDIR /var/code
COPY ./ ./

RUN \
    CGO_ENABLED=0 GOOS=linux go build \
    -o /usr/local/bin/proxy \
    ./main.go

FROM alpine:3.16.0 AS app

COPY --from=base /usr/local/bin/proxy /usr/local/bin/proxy
ENTRYPOINT ["/usr/local/bin/proxy"]
