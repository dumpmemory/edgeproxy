package proxy

import (
	"context"
	"edgeproxy/config"
	"edgeproxy/stream"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
)

type transparentProxy struct {
	ctx                      context.Context
	transparentProxyMappings []config.TransparentProxyMapping
	dialer                   Dialer
	runningListeners         []*listenerTransparentProxyMapping
}

type listenerTransparentProxyMapping struct {
	config.TransparentProxyMapping
	listener net.Listener
}

func NewTransparentProxy(ctx context.Context, dialer Dialer, transparentProxyMappings []config.TransparentProxyMapping) Proxy {
	transparentProxy := &transparentProxy{
		ctx:                      ctx,
		transparentProxyMappings: transparentProxyMappings,
		dialer:                   dialer,
	}
	return transparentProxy

}

func (t *transparentProxy) Start() {
	for _, mapping := range t.transparentProxyMappings {
		localAddr := fmt.Sprintf(":%d", mapping.ListenPort)
		log.Infof("Starting Transparent Proxy: %s", mapping.String())
		listener, err := net.Listen("tcp", localAddr)
		if err != nil {
			log.Fatalf("Error when listening port %d: %v", mapping.ListenPort, err)
		}
		listenerMapping := &listenerTransparentProxyMapping{
			TransparentProxyMapping: mapping,
			listener:                listener,
		}
		t.runningListeners = append(t.runningListeners, listenerMapping)
		go t.serve(listener, listenerMapping)
	}
}

func (s *transparentProxy) Stop() {
	for _, listenerMapping := range s.runningListeners {
		log.Infof("Stopping Transparent Proxy %s --> %s:%d", listenerMapping.listener.Addr(), listenerMapping.DestinationHost, listenerMapping.DestinationPort)
		if err := listenerMapping.listener.Close(); err != nil {
			log.Errorf("Error closing Listener %s", listenerMapping.listener.Addr().String())
		}
	}
}

func (t *transparentProxy) serve(listener net.Listener, mapping *listenerTransparentProxyMapping) {
	destinationAddr := fmt.Sprintf("%s:%d", mapping.DestinationHost, mapping.DestinationPort)
	for {
		//TODO How we stop this gracefully?
		originConn, err := listener.Accept()
		log.Debugf("Accepted new TCP Connection")
		if err != nil {
			log.Warnf("Error when accepting incoming connection %s: %v", listener.Addr().String(), err)
		}

		go t.serveConnection(originConn, listener.Addr().Network(), destinationAddr)
	}
}

func (t *transparentProxy) serveConnection(originConn net.Conn, network string, destinationAddr string) {
	defer originConn.Close()
	tunnelConn, err := t.dialer.DialContext(t.ctx, network, destinationAddr)
	if err != nil {
		log.Error(err)
	}
	defer tunnelConn.Close()
	stream.NewBidirectionalStream(tunnelConn, originConn, "tunnel", "origin").Stream()
}
