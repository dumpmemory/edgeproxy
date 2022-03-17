FROM golang:alpine3.15 as builder
RUN mkdir /app
ADD . /app/
WORKDIR /app/
RUN apk add --no-cache git gcc libc-dev && \
    go build -o /app/edgeproxy .


FROM alpine:3.15
COPY --from=builder /app/edgeproxy /edgeproxy
WORKDIR /app
RUN apk add --no-cache \
        ca-certificates \
        libmnl iproute2 iptables

ENV SERVER_PORT=9180
ENV PROXY_PORT=9080
ENV SOCKS_PORT=9022
EXPOSE $SERVER_PORT
EXPOSE $PROXY_PORT
EXPOSE $SOCKS_PORT

ENTRYPOINT ["/edgeproxy"]
CMD ["--help"] 