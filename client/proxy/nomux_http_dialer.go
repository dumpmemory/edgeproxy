package proxy

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

type httpDialer struct {
	net.Conn
	Endpoint      *url.URL
	Authenticator clientauth.Authenticator
}

func NewNoMuxHttpDialer(ctx context.Context, endpoint string, authenticator clientauth.Authenticator) (*httpDialer, error) {
	endpointUrl, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	return &httpDialer{
		Endpoint:      endpointUrl,
		Authenticator: authenticator,
	}, nil
}

func (d *httpDialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	if network == "udp" {
		return nil, fmt.Errorf("not Support %s network", network)
	}

	log.Debugf("Connecting to tunnel endpoint %s, Forwarding %s %s", d.Endpoint.String(), network, addr)
	headers := http.Header{}
	headers.Add(transport.HeaderNetworkType, transport.TcpNetType.String())
	headers.Add(transport.HeaderDstAddress, addr)
	headers.Add(transport.HeaderMuxerType, string(transport.HttpNoMuxer))
	headers.Add(transport.HeaderRouterAction, transport.ConnectionForwardRouterAction.String())
	if d.Authenticator != nil {
		d.Authenticator.AddAuthenticationHeaders(&headers)
	}
	return stream.NewHttpBiStreamConnFromEndpoint(ctx, d.Endpoint, headers)
}

func (d *httpDialer) Dial(network string, addr string) (net.Conn, error) {
	return d.DialContext(context.Background(), network, addr)
}
