package transport

import (
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"net"
	"sync"
	"time"
)

const (
	HeaderNetworkType = "X-EDGEPROXY-NETWORK"
	HeaderDstAddress  = "X-EDGEPROXY-DST"

	TCPNetwork = "tcp"
	UDPNetwork = "udp"
)

type edgeProxyReadWriter struct {
	conn       *websocket.Conn
	readMutex  sync.Mutex
	writeMutex sync.Mutex
	reader     io.Reader
}

func NewEdgeProxyReadWriter(conn *websocket.Conn) net.Conn {
	return &edgeProxyReadWriter{
		conn: conn,
	}
}

func (a *edgeProxyReadWriter) Read(b []byte) (int, error) {
	a.readMutex.Lock()
	defer a.readMutex.Unlock()

	if a.reader == nil {
		messageType, reader, err := a.conn.NextReader()
		if err != nil {
			return 0, err
		}

		if messageType != websocket.BinaryMessage {
			return 0, errors.New("unexpected websocket message type")
		}

		a.reader = reader
	}

	bytesRead, err := a.reader.Read(b)
	if err != nil {
		a.reader = nil
		if err == io.EOF {
			err = nil
		}
	}

	return bytesRead, err
}

func (a *edgeProxyReadWriter) Write(b []byte) (int, error) {
	a.writeMutex.Lock()
	defer a.writeMutex.Unlock()

	nextWriter, err := a.conn.NextWriter(websocket.BinaryMessage)
	if err != nil {
		return 0, err
	}

	bytesWritten, err := nextWriter.Write(b)
	if err != nil {
		fmt.Println("sdf")
	}
	nextWriter.Close()

	return bytesWritten, err
}

func (a *edgeProxyReadWriter) CloseWrite() error {
	a.writeMutex.Lock()
	defer a.writeMutex.Unlock()
	return a.conn.Close()
}
func (a *edgeProxyReadWriter) CloseRead() error {
	a.readMutex.Lock()
	defer a.readMutex.Unlock()
	return a.conn.Close()
}

func (a *edgeProxyReadWriter) Close() error {
	return a.conn.Close()
}

func (a *edgeProxyReadWriter) LocalAddr() net.Addr {
	return a.conn.LocalAddr()
}

func (a *edgeProxyReadWriter) RemoteAddr() net.Addr {
	return a.conn.RemoteAddr()
}

func (a *edgeProxyReadWriter) SetDeadline(t time.Time) error {
	if err := a.SetReadDeadline(t); err != nil {
		return err
	}

	return a.SetWriteDeadline(t)
}

func (a *edgeProxyReadWriter) SetReadDeadline(t time.Time) error {
	return a.conn.SetReadDeadline(t)
}

func (a *edgeProxyReadWriter) SetWriteDeadline(t time.Time) error {
	return a.conn.SetWriteDeadline(t)
}
