package proxy

import (
	"context"
	"net"
)

type Dialer interface {
	// Dial connects to the given address via the proxy.
	Dial(network, addr string) (c net.Conn, err error)
	DialContext(ctx context.Context, network, addr string) (net.Conn, error)
}

type Proxy interface {
	Start()
}
