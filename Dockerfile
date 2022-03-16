FROM alpine:3.15
WORKDIR /config
RUN apk add --no-cache \
        ca-certificates
ENV SERVER_PORT=9180
ENV PROXY_PORT=9080
ENV SOCKS_PORT=9022
EXPOSE $SERVER_PORT
EXPOSE $PROXY_PORT
EXPOSE $SOCKS_PORT

COPY edgeproxy /edgeproxy

ENTRYPOINT ["/edgeproxy"]
CMD ["--help"]
