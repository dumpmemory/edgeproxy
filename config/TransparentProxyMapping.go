package config

import "fmt"

type TransparentProxyMapping struct {
	ListenPort      int
	Network         string
	DestinationHost string
	DestinationPort int
}

func (mapping TransparentProxyMapping) String() string {
	return fmt.Sprintf("%d:%s:%s:%d", mapping.ListenPort, mapping.Network, mapping.DestinationHost, mapping.DestinationPort)
}
