FROM alpine:latest
ADD tcp-proxy-pool /tmp/tcp-proxy-pool
ENTRYPOINT ["/tmp/tcp-proxy-pool"]