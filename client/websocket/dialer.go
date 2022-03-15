package websocket

import (
	"context"
	"edgeproxy/client/clientauth"
	"edgeproxy/stream"
	"edgeproxy/transport"
	"fmt"
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

	log.Debugf("Connecting to Websocket tunnel endpoint %s, Forwarding %s %s", d.Endpoint.String(), network, addr)
	headers := http.Header{}
	headers.Add(transport.HeaderNetworkType, stream.TCPNetwork)
	headers.Add(transport.HeaderDstAddress, addr)
	headers.Add(transport.HeaderMuxerType, string(transport.HttpNoMuxer))
	headers.Add(transport.HeaderRouterAction, string(transport.ConnectionForwardRouterAction))
	if d.Authenticator != nil {
		d.Authenticator.AddAuthenticationHeaders(&headers)
	}
	return stream.NewWebsocketConnFromEndpoint(ctx, d.Endpoint, headers)
}

func (d *webSocketConnectionDialer) Dial(network string, addr string) (net.Conn, error) {
	return d.DialContext(context.Background(), network, addr)
}
