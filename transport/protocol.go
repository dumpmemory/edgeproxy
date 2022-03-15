package transport

import (
	"fmt"
)

const (
	ConnectionForwardRouterAction RouterAction = "forward"

	HeaderMuxerType    = "X-EDGEPROXY-MUXERTYPE"
	HeaderNetworkType  = "X-EDGEPROXY-NETWORK"
	HeaderRouterAction = "X-EDGEPROXY-ACTION"
	HeaderDstAddress   = "X-EDGEPROXY-DST"
)

type RouterAction string

func RouterActionFromString(routerAction string) (RouterAction, error) {
	switch routerAction {
	case string(ConnectionForwardRouterAction):
		return ConnectionForwardRouterAction, nil
	}
	return "", fmt.Errorf("router Action %s not available", routerAction)
}
