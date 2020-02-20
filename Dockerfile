FROM golang:1.13 as builder

RUN go get -v github.com/karimodm/iota-spammer

FROM debian:stable-slim

WORKDIR /app

COPY --from=builder "/go/bin/iota-spammer" "/app"

ENTRYPOINT ["/app/iota-spammer"]
