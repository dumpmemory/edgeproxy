package transport

import (
	"bytes"
	"fmt"
	"github.com/gorilla/websocket"
	"net"
	"time"
)

const (
	HeaderNetworkType = "X-EDGEPROXY-NETWORK"
	HeaderDstAddress  = "X-EDGEPROXY-DST"

	TCPNetwork = "tcp"
	UDPNetwork = "udp"
)

type edgeProxyReadWriter struct {
	*websocket.Conn
	readBuf bytes.Buffer
}

func NewEdgeProxyReadWriter(conn *websocket.Conn) *edgeProxyReadWriter {
	return &edgeProxyReadWriter{
		Conn: conn,
	}
}

func (a *edgeProxyReadWriter) Read(b []byte) (int, error) {
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

func (a *edgeProxyReadWriter) Write(b []byte) (int, error) {
	if err := a.Conn.WriteMessage(websocket.BinaryMessage, b); err != nil {
		return 0, err
	}
	return len(b), nil
}

func (a *edgeProxyReadWriter) CloseWrite() error {
	return a.Conn.Close()
}
func (a *edgeProxyReadWriter) CloseRead() error {
	return a.Conn.Close()
}

func (a *edgeProxyReadWriter) Close() error {
	return a.Conn.Close()
}

func (a *edgeProxyReadWriter) LocalAddr() net.Addr {
	return a.Conn.LocalAddr()
}

func (a *edgeProxyReadWriter) RemoteAddr() net.Addr {
	return a.Conn.RemoteAddr()
}

func (a *edgeProxyReadWriter) SetDeadline(t time.Time) error {
	if err := a.Conn.SetReadDeadline(t); err != nil {
		return fmt.Errorf("error setting read deadline: %w", err)
	}
	if err := a.Conn.SetWriteDeadline(t); err != nil {
		return fmt.Errorf("error setting write deadline: %w", err)
	}
	return nil
}
