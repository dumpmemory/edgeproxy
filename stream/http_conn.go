package stream

import (
	"context"
	"crypto/tls"
	"fmt"
	h2conn "github.com/segator/h2conn"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/http2"
	"net"
	"net/http"
	"net/url"
)

func NewHttpBiStreamConnFromEndpoint(ctx context.Context, endpoint *url.URL, headers http.Header) (net.Conn, error) {
	var conn net.Conn
	conn, err := NewHttp2BiStreamConnFromEndpoint(ctx, endpoint, headers)
	if err != nil {
		log.Debugf("Can not connect to %s via HTTP2, trying over Websocket", endpoint.String())
		conn, err = NewWebsocketConnFromEndpoint(ctx, endpoint, headers)
		if err != nil {
			return nil, err
		}
	}
	return conn, nil
}

func NewHttp2BiStreamConnFromEndpoint(ctx context.Context, endpoint *url.URL, headers http.Header) (net.Conn, error) {
	switch endpoint.Scheme {
	case "wss":
		endpoint.Scheme = "https"
		break
	case "ws":
		endpoint.Scheme = "http"
	}
	d := &h2conn.Client{
		Method: "POST",
		Header: headers,
		Client: &http.Client{
			Transport: &http2.Transport{
				AllowHTTP: true,
				DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
					return net.Dial(network, addr)
				},
			},
		},
	}

	conn, resp, err := d.Connect(ctx, endpoint.String())
	if err != nil {
		return nil, err
	}
	// Check server status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	return conn, nil
}
