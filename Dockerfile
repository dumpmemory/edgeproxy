FROM golang:alpine3.15 as builder
RUN mkdir /app
ADD . /app/
WORKDIR /app/
RUN apk add --no-cache git gcc libc-dev && \
    go build -o /app/edgeproxy .


FROM alpine:3.15
COPY --from=builder /app/edgeproxy /app/
WORKDIR /app
RUN apk add --no-cache \
        ca-certificates \
        libmnl iproute2 iptables

ENTRYPOINT ["/app/edgeproxy"]
CMD ["--help"]