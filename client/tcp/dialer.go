package tcp

import (
	"context"
	"net"
)

type dialer struct {
}

func NewTCPDialer() *dialer {
	return &dialer{}
}

func (d *dialer) Dial(network string, addr string) (net.Conn, error) {
	return net.Dial(network, addr)
}
func (d *dialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	return d.Dial(network, addr)
}
