package proxy

import (
	"context"
	"math/rand"
	"net"
)

type lbDialer struct {
	ctx     context.Context
	dialers []Dialer
}

func NewLBDialer(ctx context.Context, dialers []Dialer) *lbDialer {
	return &lbDialer{
		ctx:     ctx,
		dialers: dialers,
	}
}

func (d *lbDialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	return d.getDialer().DialContext(ctx, network, addr)
}

func (d *lbDialer) Dial(network string, addr string) (net.Conn, error) {
	return d.getDialer().Dial(network, addr)
}

func (d *lbDialer) getDialer() Dialer {
	selectedDialer := 0
	if len(d.dialers) > 1 {
		selectedDialer = rand.Intn(len(d.dialers) - 1)
	}
	return d.dialers[selectedDialer]
}
