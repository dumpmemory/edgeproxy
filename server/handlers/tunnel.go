package handlers

import (
	"context"
	"edgeproxy/server/auth"
	"edgeproxy/stream"
	"edgeproxy/transport"
	"fmt"
	h2conn "github.com/segator/h2conn"
	log "github.com/sirupsen/logrus"
	"io"
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
		var serverConn io.ReadWriteCloser
		var err error

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
			return
		}
		serverConn, err = t.tunnelConnector(res, req)
		if err != nil {
			invalidRequest(res, err)
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

//This functions translates res and req to io.ReadWriteCloser based on the Client Requested Protocol(HTTP/2 Push,HTTP1.1 Websocket)
func (t *tunnelHandler) tunnelConnector(res http.ResponseWriter, req *http.Request) (serverConn io.ReadWriteCloser, err error) {
	//Check for HTTP2 tunnel
	serverConn, err = h2conn.Accept(res, req)
	if err == nil {
		return
	}
	log.Debug("connection is not HTTP2 ready, trying WSS")
	upgradeHeader := req.Header.Get(HeaderUpgrade)
	if strings.ToLower(upgradeHeader) != "websocket" {
		return nil, fmt.Errorf("invalid bidirectional stream protocol")
	}

	serverConn, err = stream.NewWebSocketConnectFromServer(context.Background(), res, req)
	if err != nil {
		return nil, fmt.Errorf("error Initializing Websocket %v", err)
	}
	return serverConn, nil
}

func invalidRequest(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(err.Error()))
}
