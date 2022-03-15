package handlers

import (
	"context"
	"edgeproxy/server/auth"
	"edgeproxy/stream"
	"edgeproxy/transport"
	"fmt"
	"net"
	"net/http"
	"strings"
)

var (
	HeaderUpgrade = "Upgrade"
)

type httpHandlerFunc func(http.ResponseWriter, *http.Request)

type tunnelHandler struct {
}

func NewTunnelHandlder(ctx context.Context) *tunnelHandler {
	return &tunnelHandler{}
}

func (t *tunnelHandler) TunnelHandler(authenticate auth.Authenticate, authorizer auth.Authorize) httpHandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		authorized, subject := authenticate.Authenticate(res, req)
		if !authorized {
			res.WriteHeader(http.StatusUnauthorized)
			return
		}

		upgradeHeader := req.Header.Get(HeaderUpgrade)
		var serverConn net.Conn
		var err error
		//Choose bidirectional stream protocol
		if strings.ToLower(upgradeHeader) == "websocket" {
			serverConn, err = stream.NewWebSocketConnectFromServer(context.Background(), res, req)
			if err != nil {
				invalidRequest(res, fmt.Errorf("error Initializing Websocket %v", err))
				return
			}
		}

		muxerType, err := transport.MuxerTypeFromStr(req.Header.Get(transport.HeaderMuxerType))
		if err != nil {
			invalidRequest(res, err)
		}
		muxer, err := transport.NewMuxer(muxerType, req)
		if err != nil {
			invalidRequest(res, err)
		}
		router := transport.NewRouter(authorizer)
		muxer.ExecuteRouter(router, serverConn, subject.GetSubject())
	}
}

func invalidRequest(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(err.Error()))
}
