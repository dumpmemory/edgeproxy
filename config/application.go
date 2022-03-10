package config

import (
	proxy "edgeproxy/client/proxy"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

type TransportType string
type TransparentProxyMappingList []proxy.TransparentProxyMapping
type PortForwardingMappingList []proxy.PortForwardingMapping

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
		stringList = append(stringList, mapping.String())
	}
	return fmt.Sprintf("%q", stringList)
}

func (t *TransparentProxyMappingList) Set(s string) (err error) {
	//5000#TCP#1.1.1.1:5000
	var portString string
	transparentProxy := proxy.TransparentProxyMapping{}
	transparentProxyString := strings.Split(s, "#")
	transparentProxy.ListenPort, err = strconv.Atoi(transparentProxyString[0])
	if err != nil {
		return fmt.Errorf("invalid Format for Transparent Proxy Mapping: %s, %v", s, err)
	}
	transparentProxy.Network = strings.ToLower(transparentProxyString[1])
	transparentProxy.DestinationHost, portString, err = net.SplitHostPort(transparentProxyString[2])
	if err != nil {
		return fmt.Errorf("invalid Format for Transparent Proxy Mapping: %s, %v", s, err)
	}
	transparentProxy.DestinationPort, err = strconv.Atoi(portString)
	if err != nil {
		return fmt.Errorf("invalid Format for Transparent Proxy Mapping: %s, %v", s, err)
	}

	*t = append(*t, transparentProxy)
	return nil
}

func (t *TransparentProxyMappingList) Type() string {
	return "TransparentProxyMapping"
}

func (t *PortForwardingMappingList) String() string {
	var stringList []string
	for _, mapping := range *t {
		stringList = append(stringList, mapping.String())
	}
	return fmt.Sprintf("%q", stringList)
}

func (t *PortForwardingMappingList) Set(s string) (err error) {
	//5000#TCP#wss://myendpoint:443

	portForwardingMapping := proxy.PortForwardingMapping{}
	transparentProxyString := strings.Split(s, "#")
	portForwardingMapping.ListenPort, err = strconv.Atoi(transparentProxyString[0])
	if err != nil {
		return fmt.Errorf("invalid Format for Transparent Proxy Mapping: %s, %v", s, err)
	}
	portForwardingMapping.Network = strings.ToLower(transparentProxyString[1])
	portForwardingMapping.Endpoint, err = url.Parse(transparentProxyString[2])
	if err != nil {
		return fmt.Errorf("invalid Endpoint Format for Port Forwarding Mapping: %s, %v", s, err)
	}

	*t = append(*t, portForwardingMapping)
	return nil
}

func (t *PortForwardingMappingList) Type() string {
	return "PortForwardingMapping"
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
	PortForwardList          PortForwardingMappingList
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
	if c.TransportType == WebsocketTransport && (c.EnableProxy || c.EnableSocks5) {
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
