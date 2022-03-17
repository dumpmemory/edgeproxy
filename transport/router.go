package transport

import (
	"edgeproxy/metrics"
	"edgeproxy/server/auth"
	"edgeproxy/stream"
	"fmt"
	"net"
)

type Router struct {
	authorizer auth.Authorize
}

func NewRouter(authorizer auth.Authorize) *Router {
	return &Router{
		authorizer: authorizer,
	}
}

func (r *Router) ConnectionForward(sourceConn net.Conn, forward auth.ForwardAction) error {
	if policyRes := r.authorizer.AuthorizeForward(forward); policyRes {
		metrics.IncrementRouterForwardAcceptedConnections()
		dstConn, err := net.Dial(forward.NetType, forward.DestinationAddr)
		if err != nil {
			return fmt.Errorf("can not connect to %s: %v", forward.DestinationAddr, err)
		}
		defer dstConn.Close()

		stream.NewBidirectionalStream(sourceConn, dstConn, "tunnel", "destination").Stream()
		return nil
	} else {
		return fmt.Errorf("denied %s access to %s/%s", forward.Subject, forward.NetType, forward.DestinationAddr)
	}
}
