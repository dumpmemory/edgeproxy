package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type TransportType string
type TransparentProxyMappingList []TransparentProxyMapping
type PortForwardingMappingList []PortForwardingMapping

const (
	WebsocketTransport    TransportType = "WebSocketTransport"
	WebsocketMuxTransport TransportType = "WebSocketMuxTransport"
	TcpTransport          TransportType = "TcpTransport"
	WireguardTransport    TransportType = "WireguardTransport"
	UdpTransport          TransportType = "UDPTransport"
	QuickTransport        TransportType = "QUICKTransport"
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
	transparentProxy := TransparentProxyMapping{}
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

	portForwardingMapping := PortForwardingMapping{}
	transparentProxyString := strings.Split(s, "#")
	portForwardingMapping.ListenPort, err = strconv.Atoi(transparentProxyString[0])
	if err != nil {
		return fmt.Errorf("invalid Format for Transparent Proxy Mapping: %s, %v", s, err)
	}
	portForwardingMapping.Network = strings.ToLower(transparentProxyString[1])
	portForwardingMapping.Endpoint, err = url.Parse(transparentProxyString[2])
	if err != nil {
		return fmt.Errorf("invalid endpoint Format for Port Forwarding Mapping: %s, %v", s, err)
	}

	*t = append(*t, portForwardingMapping)
	return nil
}

func (t *PortForwardingMappingList) Type() string {
	return "PortForwardingMapping"
}

type ApplicationConfig struct {
	ClientConfig *ClientConfig `mapstructure:"client"`
	ServerConfig *ServerConfig `mapstructure:"server"`
}

type ClientConfig struct {
	EnableProxy              bool `mapstructure:"proxy"`
	EnableSocks5             bool `mapstructure:"socks5"`
	HttpProxyPort            int
	Socks5Port               int
	TransportType            TransportType
	WebSocketTransportConfig WebSocketTransportConfig
	TransparentProxyList     TransparentProxyMappingList
	PortForwardList          PortForwardingMappingList
	Auth                     ClientAuthConfig `mapstructure:"clientauth"`
}

type ServerConfig struct {
	HttpPort int              `mapstructure:"httpPort"`
	Auth     ServerAuthConfig `mapstructure:"clientauth"`
}

type ClientAuthConfig struct {
	CaConfig ClientAuthCaConfig `mapstructure:"ca"`
}

type ServerAuthConfig struct {
	CaConfig      ServerAuthCaConfig `mapstructure:"ca"`
	AclPolicyPath AclCollection      `mapstructure:"acl"`
}
type AclCollection struct {
	IpPath     string `mapstructure:"ip"`
	DomainPath string `mapstructure:"domain"`
}

type PathsConfig struct {
	Allowed []string `mapstructure:"allowed"`
	Denied  []string `mapstructure:"denied"`
}

func (c PathsConfig) AllowedPath(path string) bool {
	for _, allow := range c.Allowed {
		// TODO: could pre-compile these regexes
		r, _ := regexp.Compile(allow)
		if r.MatchString(path) {
			for _, deny := range c.Denied {
				rDeny, _ := regexp.Compile(deny)
				if rDeny.MatchString(path) {
					return false
				}
			}
			return true
		}
	}
	return false
}

type ClientAuthCaConfig struct {
	Key         string `mapstructure:"key"`
	Certificate string `mapstructure:"cert"`
}

type ServerAuthCaConfig struct {
	TrustedRoot      string `mapstructure:"root_bundle"`
	SpireTrustDomain string `mapstructure:"trust_domain"`
	Paths            PathsConfig
}

func (s ServerConfig) Validate() error {
	if s.HttpPort <= 0 || s.HttpPort > 65635 {
		return fmt.Errorf("invalid Server Http port %d", s.HttpPort)
	}
	if s.Auth.CaConfig.TrustedRoot != "" {
		if s.Auth.CaConfig.SpireTrustDomain == "" {
			return errors.New("must set a SPIFFE trust domain")
		}
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
		return fmt.Errorf("invalid WebSocket Tunnel endpoint %v", err)
	}
	return nil
}
