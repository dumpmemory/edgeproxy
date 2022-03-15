package transport

import (
	"fmt"
	"net"
	"net/http"
)

const (
	HttpNoMuxer MuxerType = "httpNoMuxer"
	YamuxMuxer  MuxerType = "yamuxMuxer"
)

type Muxer interface {
	ExecuteRouter(router *Router, tunnelConn net.Conn, subject string)
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
