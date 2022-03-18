package transport

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
)

const (
	HttpNoMuxer MuxerType = "httpNoMuxer"
	YamuxMuxer  MuxerType = "yamuxMuxer"
)

type Dialer interface {
	// Dial connects to the given address via the proxy.
	Dial(network, addr string) (c net.Conn, err error)
	DialContext(ctx context.Context, network, addr string) (net.Conn, error)
}

type Muxer interface {
	ExecuteServerRouter(router *Router, tunnelConn io.ReadWriteCloser, subject string) error
}
type MuxerType string

func MuxerTypeFromStr(muxerTypeStr string) (MuxerType, error) {
	switch muxerTypeStr {
	case string(HttpNoMuxer):
		return HttpNoMuxer, nil
	case string(YamuxMuxer):
		return YamuxMuxer, nil
	}
	return "", fmt.Errorf("yamuxMuxer Type %s not available", muxerTypeStr)
}

func NewMuxer(muxerType MuxerType, r *http.Request) (Muxer, error) {
	switch muxerType {
	case HttpNoMuxer:
		return NewHttpNoMuxer(r)
	case YamuxMuxer:
		return NewYamuxMuxer()
	}
	return nil, fmt.Errorf("no Muxer Found %s", muxerType)
}
