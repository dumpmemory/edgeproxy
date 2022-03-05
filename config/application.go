package config

import (
	"fmt"
	proxy "httpProxy/client/proxy"
	"net/url"
	"strconv"
	"strings"
)

type TransportType string
type TransparentProxyMappingList []proxy.TransparentProxyMapping

const (
	WebsocketTransport TransportType = "WebSocketTransport"
	TcpTransport       TransportType = "TcpTransport"
	WireguardTransport TransportType = "WireguardTransport"
	UdpTransport       TransportType = "UDPTransport"
	QuickTransport     TransportType = "QUICKTransport"
)

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

func (t *TransparentProxyMappingList) String() string {
	var stringList []string
	for _, mapping := range *t {
		stringList = append(stringList, fmt.Sprintf("%d:%s:%s:%d", mapping.ListenPort, mapping.Network, mapping.DestinationHost, mapping.DestinationPort))
	}
	return fmt.Sprintf("%q", stringList)
}

func (t *TransparentProxyMappingList) Set(s string) (err error) {
	//5000:TCP:1.1.1.1:5000

	transparentProxy := proxy.TransparentProxyMapping{}
	transparentProxyString := strings.Split(s, ":")
	transparentProxy.ListenPort, err = strconv.Atoi(transparentProxyString[0])
	if err != nil {
		return fmt.Errorf("invalid Format for Transparent Proxy Mapping: %s, %v", s, err)
	}
	transparentProxy.Network = transparentProxyString[1]
	transparentProxy.DestinationHost = transparentProxyString[2]
	transparentProxy.DestinationPort, err = strconv.Atoi(transparentProxyString[3])
	if err != nil {
		return fmt.Errorf("invalid Format for Transparent Proxy Mapping: %s, %v", s, err)
	}
	*t = append(*t, transparentProxy)
	return nil
}

func (t *TransparentProxyMappingList) Type() string {
	return "TransparentProxyMapping"
}

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
	TransparentProxyList     TransparentProxyMappingList
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
