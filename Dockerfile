ARG GO_VERSION=1.23.2
ARG COREDNS_VERSION=1.11.3

FROM golang:${GO_VERSION}-alpine as builder

ARG COREDNS_VERSION

RUN apk add --no-cache git make

WORKDIR /coredns

RUN git clone --depth 1 --branch v${COREDNS_VERSION} https://github.com/coredns/coredns.git .

RUN sed -i '/cache:cache/i blocker:github.com/GustavoKatel/coredns-blocker' plugin.cfg

ADD . /coredns/plugin/blocker

RUN echo "replace github.com/GustavoKatel/coredns-blocker => /coredns/plugin/blocker" >> go.mod
RUN go get github.com/GustavoKatel/coredns-blocker

RUN go generate && make

FROM alpine:latest

COPY --from=builder /coredns/coredns /coredns/coredns

ENTRYPOINT ["/coredns/coredns"]
