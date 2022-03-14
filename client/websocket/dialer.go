package websocket

import (
	"context"
	"edgeproxy/client/clientauth"
	"edgeproxy/transport"
	"fmt"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
	"net/url"
)

type webSocketConnectionDialer struct {
	net.Conn
	Endpoint      *url.URL
	Authenticator clientauth.Authenticator
}

func NewWebSocketDialer(endpoint string, authenticator clientauth.Authenticator) (*webSocketConnectionDialer, error) {
	endpointUrl, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	return &webSocketConnectionDialer{
		Endpoint:      endpointUrl,
		Authenticator: authenticator,
	}, nil
}

func (d *webSocketConnectionDialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	if network == "udp" {
		return nil, fmt.Errorf("not Support %s network", network)
	}
	switch d.Endpoint.Scheme {
	case "https":
		d.Endpoint.Scheme = "wss"
		break
	case "http":
		d.Endpoint.Scheme = "ws"
	}
	log.Debugf("Connecting to Websocket tunnel endpoint %s, Forwarding %s %s", d.Endpoint.String(), network, addr)
	headers := http.Header{}
	headers.Add(transport.HeaderNetworkType, transport.TCPNetwork)
	headers.Add(transport.HeaderDstAddress, addr)
	if d.Authenticator != nil {
		d.Authenticator.AddAuthenticationHeaders(&headers)
	}
	wssCon, _, err := websocket.DefaultDialer.DialContext(ctx, d.Endpoint.String(), headers)
	if err != nil {
		return nil, fmt.Errorf("error when dialing Websocket tunnel %s: %v", d.Endpoint.String(), err)
	}
	edgeReadWriter := transport.NewEdgeProxyReadWriter(wssCon)
	return edgeReadWriter, nil
}

func (d *webSocketConnectionDialer) Dial(network string, addr string) (net.Conn, error) {
	return d.DialContext(context.Background(), network, addr)
}
