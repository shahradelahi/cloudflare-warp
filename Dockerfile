ARG ALPINE_VERSION=latest

FROM --platform=$BUILDPLATFORM alpine:${ALPINE_VERSION} AS alpine
ENV TZ=Etc/UTC
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ >/etc/timezone
RUN apk update \
 && apk add -U --no-cache \
  ca-certificates \
 && rm -rf /var/cache/apk/*

FROM golang:alpine AS builder

WORKDIR /src
COPY . /src

ENV GOCACHE=/gocache
RUN --mount=type=cache,target="/gocache" apk add --update --no-cache make git \
    && make cloudflare-warp

FROM alpine
LABEL org.opencontainers.image.source="https://github.com/shahradelahi/cloudflare-warp"

# Create and set permissions for the data directory
RUN mkdir -p /var/lib/cloudflare-warp

COPY docker/entrypoint.sh /entrypoint.sh
COPY --from=builder /src/build/warp /usr/bin/warp

RUN apk add --update --no-cache iptables iproute2 tzdata \
    && chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
