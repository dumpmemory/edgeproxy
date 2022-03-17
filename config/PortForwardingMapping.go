package config

import (
	"fmt"
	"net/url"
)

type PortForwardingMapping struct {
	ListenPort int
	Network    string
	Endpoint   *url.URL
}

func (mapping PortForwardingMapping) String() string {
	return fmt.Sprintf("%d:%s:%s", mapping.ListenPort, mapping.Network, mapping.Endpoint.String())
}
