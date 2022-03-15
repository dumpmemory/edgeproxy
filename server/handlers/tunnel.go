package handlers

import (
	"context"
	"edgeproxy/server/auth"
	"edgeproxy/transport"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
)

var (
	connectionsAccepted = promauto.NewCounter(prometheus.CounterOpts{
		Name: "edgeproxy_wss_connections_accepted",
		Help: "Accepted WSS connections",
	})
	tunnelreadProxiedKBytes = promauto.NewCounter(prometheus.CounterOpts{
		Name: "edgeproxy_tunnel_read_kilobytes",
		Help: "Number of read bytes in the tunnel",
	})
	tunnelwrittenProxiedKBytes = promauto.NewCounter(prometheus.CounterOpts{
		Name: "edgeproxy_tunnel_written_kilobytes",
		Help: "Number of Written bytes in the tunnel",
	})
)

type httpHandlerFunc func(http.ResponseWriter, *http.Request)

type tunnelHandler struct {
}

func NewTunnelHandlder(ctx context.Context) *tunnelHandler {
	return &tunnelHandler{}
}

func (t *tunnelHandler) TunnelHandler(authenticate auth.Authenticate, authorizer auth.Authorize) httpHandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		authorized, subject := authenticate.Authenticate(res, req)
		if !authorized {
			res.WriteHeader(http.StatusUnauthorized)
			return
		}

		netType := req.Header.Get(transport.HeaderNetworkType)
		dstAddr := req.Header.Get(transport.HeaderDstAddress)
		if netType == "" {
			invalidRequest(res, fmt.Errorf("invalid Net Type"))
			return
		}
		if dstAddr == "" {
			invalidRequest(res, fmt.Errorf("invalid dst Addr"))
			return
		}
		if subject != nil {
			// run this subject and the requested network action through casbin
			subj := subject.GetSubject()
			forwardAction := auth.NewForwardAction(subject, dstAddr, netType)
			if policyRes := authorizer.AuthorizeForward(forwardAction); policyRes {
				//TODO support http2, if is available prefered over WebSocket
				webSocketStream(res, req, dstAddr, netType)
				return
			} else {
				log.Infof("denied %s access to %s/%s", subj, netType, dstAddr)
				res.WriteHeader(http.StatusForbidden)
				return
			}
		} else {
			log.Infof("no subject, denied access to %s/%s", netType, dstAddr)
			res.WriteHeader(http.StatusForbidden)
			return
		}
	}
}

func invalidRequest(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(err.Error()))
}

func webSocketStream(res http.ResponseWriter, req *http.Request, dstAddr, netType string) error {
	tunnelConn, err := transport.NewWebSocketConnectFromServer(context.Background(), res, req)
	if err != nil {
		return err
	}
	dstConn, err := net.Dial(netType, dstAddr)
	if err != nil {
		return fmt.Errorf("can not connect to %s: %v", dstAddr, err)
	}
	defer dstConn.Close()
	connectionsAccepted.Inc()
	readBytes, writtenBytes := transport.NewBidirectionalStream(tunnelConn, dstConn, "tunnel", "origin").Stream()
	tunnelwrittenProxiedKBytes.Add(float64(writtenBytes) / 1024.0)
	tunnelreadProxiedKBytes.Add(float64(readBytes) / 1024.0)
	return nil
}
