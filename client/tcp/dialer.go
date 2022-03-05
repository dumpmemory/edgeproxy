package tcp

import (
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
