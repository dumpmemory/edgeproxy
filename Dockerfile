

FROM alpine:3.15
WORKDIR /app
RUN apk add --no-cache \
        ca-certificates \
        libmnl iproute2 iptables \
ENV PORT=9180
EXPOSE $PORT
COPY edgeproxy /

ENTRYPOINT ["/edgeproxy"]
CMD ["--help"]