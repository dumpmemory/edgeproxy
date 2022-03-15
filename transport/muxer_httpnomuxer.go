package transport

import (
	"edgeproxy/server/auth"
	"fmt"
	"net"
	"net/http"
)

type httpNoMuxer struct {
	netType      string
	dstAddr      string
	routerAction RouterAction
}

func NewHttpNoMuxer(req *http.Request) (*httpNoMuxer, error) {
	netType := req.Header.Get(HeaderNetworkType)
	dstAddr := req.Header.Get(HeaderDstAddress)
	routerAction, err := RouterActionFromString(req.Header.Get(HeaderRouterAction))
	if err != nil {
		return nil, err
	}
	if netType == "" {
		return nil, fmt.Errorf("invalid Net Type")
	}
	if dstAddr == "" {
		return nil, fmt.Errorf("invalid dst Addr")
	}

	directRouter := &httpNoMuxer{
		netType:      netType,
		dstAddr:      dstAddr,
		routerAction: routerAction,
	}

	return directRouter, nil
}

func (h *httpNoMuxer) ExecuteRouter(router *Router, tunnelConn net.Conn, subject string) {
	switch h.routerAction {
	case ConnectionForwardRouterAction:
		router.ConnectionForward(tunnelConn, auth.NewForwardAction(subject, h.dstAddr, h.netType))
	}
}
