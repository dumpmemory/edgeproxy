package server

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"httpProxy/transport"
	"io"
	"net"
	"net/http"
	"sync"
)

type wsHandler struct {
	upgrader websocket.Upgrader
	ctx      context.Context
}

type halfClosable interface {
	net.Conn
	CloseWrite() error
	CloseRead() error
}

func NewWebSocketHandler(ctx context.Context) WebSocketHandler {
	return &wsHandler{
		ctx:      ctx,
		upgrader: websocket.Upgrader{},
	}
}

func copyOrWarn(dst io.Writer, src io.Reader, wg *sync.WaitGroup) {
	if _, err := io.Copy(dst, src); err != nil {
		log.Warnf("Error copying to client: %s", err)
	}
	wg.Done()
}

func copyAndClose(dst, src halfClosable) {
	if _, err := io.Copy(dst, src); err != nil {
		log.Warnf("Error copying to client: %s", err)
	}
	dst.CloseWrite()
	src.CloseRead()
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
		log.Errorf("Error during connection upgrade: %v", err)
	}
	edgeReadWriter := transport.NewEdgeProxyReadWriter(wsConn)

	backendConn, err := net.Dial(netType, dstAddr)
	if err != nil {
		log.Errorf("Can not connect to %s: %v", dstAddr, err)
		return
	}
	targetTCP, targetOK := backendConn.(halfClosable)
	proxyClientTCP, clientOK := edgeReadWriter.(halfClosable)
	if targetOK && clientOK {
		go copyAndClose(targetTCP, proxyClientTCP)
		go copyAndClose(proxyClientTCP, targetTCP)
	} else {
		go func() {
			var wg sync.WaitGroup
			wg.Add(2)
			go copyOrWarn(backendConn, edgeReadWriter, &wg)
			go copyOrWarn(edgeReadWriter, backendConn, &wg)
			wg.Wait()
			edgeReadWriter.Close()
			backendConn.Close()
		}()
	}
}

func (ws *wsHandler) InvalidRequest(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusBadGateway)
	w.Write([]byte(err.Error()))
}
