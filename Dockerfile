FROM golang:alpine as builder

ENV GO111MODULE=on

COPY . /go/src/github.com/Luzifer/cloudbox
WORKDIR /go/src/github.com/Luzifer/cloudbox

RUN set -ex \
 && apk add --update \
      build-base \
      git \
      sqlite \
 && go install -ldflags "-X main.version=$(git describe --tags --always || echo dev)" \
      github.com/Luzifer/cloudbox/cmd/cloudbox


FROM alpine:latest

LABEL maintainer "Knut Ahlers <knut@ahlers.me>"

RUN set -ex \
 && apk --no-cache add \
      ca-certificates \
      sqlite

COPY --from=builder /go/bin/cloudbox /usr/local/bin/cloudbox

ENTRYPOINT ["/usr/local/bin/cloudbox"]
CMD ["--"]

# vim: set ft=Dockerfile:
