package transport

import "net"

type yamuxMuxer struct {
}

func NewYamuxMuxer() (*yamuxMuxer, error) {
	m := &yamuxMuxer{}
	return m, nil
}

func (h *yamuxMuxer) ExecuteRouter(router *Router, tunnelConn net.Conn, subject string) {
	panic("not implemented")
}
