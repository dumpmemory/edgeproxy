package proxy

import (
	"context"
	"edgeproxy/transport"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
	"net/url"
)

type portForwarding struct {
	ctx                   context.Context
	portForwardingMapping []PortForwardingMapping
	dialer                Dialer
	runningListeners      []*listenerPortForwardingMapping
}

type PortForwardingMapping struct {
	ListenPort int
	Network    string
	Endpoint   *url.URL
}

func (mapping PortForwardingMapping) String() string {
	return fmt.Sprintf("%d:%s:%s", mapping.ListenPort, mapping.Network, mapping.Endpoint.String())
}

type listenerPortForwardingMapping struct {
	PortForwardingMapping
	listener net.Listener
}

func NewPortForwarding(ctx context.Context, dialer Dialer, portForwardingMapping []PortForwardingMapping) *portForwarding {
	portForwarding := &portForwarding{
		ctx:                   ctx,
		portForwardingMapping: portForwardingMapping,
		dialer:                dialer,
	}
	return portForwarding

}

func (t *portForwarding) Start() {
	for _, mapping := range t.portForwardingMapping {
		localAddr := fmt.Sprintf(":%d", mapping.ListenPort)
		log.Infof("Starting Port Forwarding: %s", mapping.String())
		listener, err := net.Listen(mapping.Network, localAddr)
		if err != nil {
			log.Fatalf("Error when listening port %d: %v", mapping.ListenPort, err)
		}
		listenerMapping := &listenerPortForwardingMapping{
			PortForwardingMapping: mapping,
			listener:              listener,
		}
		t.runningListeners = append(t.runningListeners, listenerMapping)
		go t.acceptSocketConnection(listener, listenerMapping.Endpoint)
	}
}

func (s *portForwarding) Stop() {
	for _, listenerMapping := range s.runningListeners {
		log.Infof("Stopping Port Forwarding %s", listenerMapping.listener.Addr())
		if err := listenerMapping.listener.Close(); err != nil {
			log.Errorf("Error closing Listener %s", listenerMapping.listener.Addr().String())
		}
	}
}

func (t *portForwarding) acceptSocketConnection(listener net.Listener, endpoint *url.URL) {
	for {
		//TODO How we stop this gracefully?
		fd, err := listener.Accept()
		if err != nil {
			log.Warnf("Error when accepting incoming connection %s: %v", listener.Addr().String(), err)
		}
		go t.handleSocketConnection(fd, endpoint)
	}
}

func (t *portForwarding) handleSocketConnection(originConn net.Conn, endpoint *url.URL) error {
	defer originConn.Close()
	tunnelConn, err := transport.NewWebsocketConnFromEndpoint(context.Background(), endpoint, nil)
	if err != nil {
		return err
	}
	defer tunnelConn.Close()
	transport.NewBidirectionalStream(tunnelConn, originConn, "tunnel", "origin").Stream()
	return nil
}

func closeRespBody(resp *http.Response) {
	if resp != nil {
		_ = resp.Body.Close()
	}
}
