FROM golang:alpine AS builder

WORKDIR /src
COPY . /src

RUN apk add --update --no-cache make git \
    && make cloudflare-warp

FROM alpine:latest
LABEL org.opencontainers.image.source="https://github.com/shahradelahi/cloudflare-warp"

# Create and set permissions for the data directory
RUN mkdir -p /var/lib/cloudflare-warp

COPY docker/entrypoint.sh /entrypoint.sh
COPY --from=builder /src/build/warp /usr/bin/warp

RUN apk add --update --no-cache iptables iproute2 tzdata \
    && chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
