package proxy

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"httpProxy/transport"
	"net"
)

type transparentProxy struct {
	ctx                      context.Context
	transparentProxyMappings []TransparentProxyMapping
	dialer                   Dialer
	runningListeners         []*listenerTransparentProxyMapping
}

type TransparentProxyMapping struct {
	ListenPort      int
	Network         string
	DestinationHost string
	DestinationPort int
}

type listenerTransparentProxyMapping struct {
	TransparentProxyMapping
	listener net.Listener
}

func NewTransParentProxy(ctx context.Context, dialer Dialer, transparentProxyMappings []TransparentProxyMapping) Proxy {
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
		log.Infof("Starting Transparent Proxy: PROTO %s  %s --> %s:%d", mapping.Network, localAddr, mapping.DestinationHost, mapping.DestinationPort)
		listener, err := net.Listen("tcp", localAddr)
		if err != nil {
			log.Fatalf("Error when listening port %d: %v", mapping.ListenPort, err)
		}
		listenerMapping := &listenerTransparentProxyMapping{
			TransparentProxyMapping: mapping,
			listener:                listener,
		}
		t.runningListeners = append(t.runningListeners, listenerMapping)
		go t.proxyHost(listener, listenerMapping)
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

func (t *transparentProxy) proxyHost(listener net.Listener, mapping *listenerTransparentProxyMapping) {
	destinationAddr := fmt.Sprintf("%s:%d", mapping.DestinationHost, mapping.DestinationPort)
	for {
		//TODO How we stop this gracefully?
		fd, err := listener.Accept()
		if err != nil {
			log.Warnf("Error when accepting incoming connection %s: %v", listener.Addr().String(), err)
		}
		proxyConn, err := t.dialer.DialContext(t.ctx, listener.Addr().Network(), destinationAddr)
		if err != nil {
			log.Error(err)
			continue
		}
		transport.ProxyConnection(fd, proxyConn)
	}
}
