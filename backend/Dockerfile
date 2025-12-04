FROM golang:1.25-alpine AS builder

RUN apk update \
  && apk upgrade --no-cache \
  && apk add --no-cache git ca-certificates \
  && update-ca-certificates

ENV GOPRIVATE=github.com/X/*

WORKDIR /usr/src/app

COPY . .
RUN --mount=type=secret,id=git_token \
    git config --global url."https://$(cat /run/secrets/git_token):x-oauth-basic@github.com/".insteadOf "https://github.com/" \
    && go mod download \
    && go mod verify

RUN GOEXPERIMENT=jsonv2 CGO_ENABLED=0 GOOS=linux go build -a -ldflags="-s -w" -installsuffix cgo -o /usr/src/bin/app

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/src/bin/app .
USER 1000

ENTRYPOINT ["./app"]
