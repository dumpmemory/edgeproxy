package stream

import (
	"bytes"
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"net"
	"net/http"
	"net/url"
	"time"
)

const (
	TCPNetwork = "tcp"
)

type websocketReadWriter struct {
	ctx context.Context
	*websocket.Conn
	readBuf bytes.Buffer
}

func NewWebsocketConnFromEndpoint(ctx context.Context, endpoint *url.URL, headers http.Header) (*websocketReadWriter, error) {
	switch endpoint.Scheme {
	case "https":
		endpoint.Scheme = "wss"
		break
	case "http":
		endpoint.Scheme = "ws"
	}
	wssDialer := websocket.Dialer{
		HandshakeTimeout: 45 * time.Second,
		ReadBufferSize:   32768,
		WriteBufferSize:  32768,
	}
	wssCon, _, err := wssDialer.DialContext(ctx, endpoint.String(), headers)
	if err != nil {
		return nil, fmt.Errorf("error when dialing Websocket tunnel %s: %v", endpoint, err)
	}
	return &websocketReadWriter{
		ctx:  ctx,
		Conn: wssCon,
	}, nil
}

func NewWebSocketConnectFromServer(ctx context.Context, res http.ResponseWriter, req *http.Request) (*websocketReadWriter, error) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  32768,
		WriteBufferSize: 32768,
	}
	wsConn, err := upgrader.Upgrade(res, req, nil)
	if err != nil {
		return nil, err
	}
	return &websocketReadWriter{
		ctx:  ctx,
		Conn: wsConn,
	}, nil
}

func (a *websocketReadWriter) Read(b []byte) (int, error) {
	if a.readBuf.Len() > 0 {
		return a.readBuf.Read(b)
	}
	_, message, err := a.Conn.ReadMessage()
	if err != nil {
		return 0, err
	}
	copied := copy(b, message)
	a.readBuf.Write(message[copied:])
	return copied, nil
}

func (a *websocketReadWriter) Write(b []byte) (int, error) {
	if err := a.Conn.WriteMessage(websocket.BinaryMessage, b); err != nil {
		return 0, err
	}
	return len(b), nil
}

func (a *websocketReadWriter) CloseWrite() error {
	return a.Conn.Close()
}
func (a *websocketReadWriter) CloseRead() error {
	return a.Conn.Close()
}

func (a *websocketReadWriter) Close() error {
	return a.Conn.Close()
}

func (a *websocketReadWriter) LocalAddr() net.Addr {
	return a.Conn.LocalAddr()
}

func (a *websocketReadWriter) RemoteAddr() net.Addr {
	return a.Conn.RemoteAddr()
}

func (a *websocketReadWriter) SetDeadline(t time.Time) error {
	if err := a.Conn.SetReadDeadline(t); err != nil {
		return fmt.Errorf("error setting read deadline: %w", err)
	}
	if err := a.Conn.SetWriteDeadline(t); err != nil {
		return fmt.Errorf("error setting write deadline: %w", err)
	}
	return nil
}
