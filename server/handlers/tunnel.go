package handlers

import (
	"context"
	"edgeproxy/server/auth"
	"edgeproxy/stream"
	"edgeproxy/transport"
	"fmt"
	log "github.com/sirupsen/logrus"
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

		muxerType, err := transport.MuxerTypeFromStr(req.Header.Get(transport.HeaderMuxerType))
		if err != nil {
			invalidRequest(res, err)
			return
		}
		muxer, err := transport.NewMuxer(muxerType, req)
		if err != nil {
			invalidRequest(res, err)
		}

		var serverConn net.Conn
		upgradeHeader := req.Header.Get(HeaderUpgrade)

		//Choose bidirectional stream protocol
		if strings.ToLower(upgradeHeader) == "websocket" {
			serverConn, err = stream.NewWebSocketConnectFromServer(context.Background(), res, req)
			if err != nil {
				invalidRequest(res, fmt.Errorf("error Initializing Websocket %v", err))
				return
			}
		} else {
			invalidRequest(res, fmt.Errorf("invalid bidirectioanl stream protocol"))
			return
		}

		router := transport.NewRouter(authorizer)
		err = muxer.ExecuteServerRouter(router, serverConn, subject.GetSubject())
		if err != nil {
			log.Debug(err)
			//invalidRequest(res, err)
		}
	}
}

func invalidRequest(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(err.Error()))
}
