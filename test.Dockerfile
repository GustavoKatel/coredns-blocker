ARG GO_VERSION=1.21
ARG COREDNS_VERSION=1.11.1

FROM golang:${GO_VERSION}-alpine as builder

ARG COREDNS_VERSION

RUN apk add --no-cache git make

WORKDIR /coredns

RUN git clone https://github.com/coredns/coredns.git .

RUN git checkout v${COREDNS_VERSION}

RUN sed -i '/cache:cache/i blocker:blocker' plugin.cfg
