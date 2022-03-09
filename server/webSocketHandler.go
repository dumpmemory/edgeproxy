package server

import (
	"context"
	"edgeProxy/transport"
	"fmt"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
)

type wsHandler struct {
	upgrader websocket.Upgrader
	ctx      context.Context
}

func NewWebSocketHandler(ctx context.Context) WebSocketHandler {
	return &wsHandler{
		ctx:      ctx,
		upgrader: websocket.Upgrader{},
	}
}

func (ws *wsHandler) socketHandler(w http.ResponseWriter, r *http.Request) {
	netType := r.Header.Get(transport.HeaderNetworkType)
	dstAddr := r.Header.Get(transport.HeaderDstAddress)
	if netType == "" {
		ws.InvalidRequest(w, fmt.Errorf("invalid Net Type"))
		return
	}
	if dstAddr == "" {
		ws.InvalidRequest(w, fmt.Errorf("invalid dst Addr"))
		return
	}

	wsConn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		ws.InvalidRequest(w, fmt.Errorf("error during connection upgrade: %v", err))
		return
	}
	tunnelConn := transport.NewEdgeProxyReadWriter(wsConn)
	dstConn, err := net.Dial(netType, dstAddr)
	if err != nil {
		log.Errorf("Can not connect to %s: %v", dstAddr, err)
		return
	}
	defer dstConn.Close()
	transport.Stream(tunnelConn, dstConn)

}

func (ws *wsHandler) InvalidRequest(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(err.Error()))
}
