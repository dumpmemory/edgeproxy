package transport

import (
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"sync"
)

type halfClosable interface {
	net.Conn
	CloseWrite() error
	CloseRead() error
}

func ProxyConnection(conn1 net.Conn, conn2 net.Conn) {
	targetTCP, targetOK := conn1.(halfClosable)
	proxyClientTCP, clientOK := conn2.(halfClosable)
	if targetOK && clientOK {
		go copyAndClose(targetTCP, proxyClientTCP)
		go copyAndClose(proxyClientTCP, targetTCP)
	} else {
		go func() {
			var wg sync.WaitGroup
			wg.Add(2)
			go copyOrWarn(conn1, conn2, &wg)
			go copyOrWarn(conn2, conn1, &wg)
			wg.Wait()
			conn1.Close()
			conn2.Close()
		}()
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
