package config

import (
	"fmt"
	"net/url"
)

type TransportType string

func (t *TransportType) String() string {
	return string(*t)
}

func (t *TransportType) Set(s string) error {
	*t = TransportType(s)
	return nil
}

func (t TransportType) Type() string {
	return "TransportType"
}

const (
	WebsocketTransport TransportType = "WebSocketTransport"
	TcpTransport       TransportType = "TcpTransport"
	WireguardTransport TransportType = "WireguardTransport"
	UdpTransport       TransportType = "UDPTransport"
	QuickTransport     TransportType = "QUICKTransport"
)

type ApplicationConfig struct {
	ClientConfig ClientConfig
	ServerConfig ServerConfig
}

type ClientConfig struct {
	EnableProxy              bool
	EnableSocks5             bool
	HttpProxyPort            int
	Socks5Port               int
	TransportType            TransportType
	WebSocketTransportConfig WebSocketTransportConfig
}

type ServerConfig struct {
	HttpPort int
}

func (s ServerConfig) Validate() error {
	if s.HttpPort <= 0 || s.HttpPort > 65635 {
		return fmt.Errorf("invalid Server Http port %d", s.HttpPort)
	}
	return nil
}

type WebSocketTransportConfig struct {
	WebSocketTunnelEndpoint string
}

func (c ClientConfig) Validate() (err error) {
	if c.TransportType == WebsocketTransport {
		if err = c.WebSocketTransportConfig.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (c WebSocketTransportConfig) Validate() error {
	if len(c.WebSocketTunnelEndpoint) == 0 {
		return fmt.Errorf("WebSocketTunnelEndpoint is mandatory")
	}
	if _, err := url.Parse(c.WebSocketTunnelEndpoint); err != nil {
		return fmt.Errorf("invalid WebSocket Tunnel Endpoint %v", err)
	}
	return nil
}
