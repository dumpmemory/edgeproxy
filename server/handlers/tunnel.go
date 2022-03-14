package handlers

import (
	"context"
	"edgeproxy/server/auth"
	"edgeproxy/transport"
	"fmt"
	"github.com/gorilla/websocket"
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
)

type httpHandlerFunc func(http.ResponseWriter, *http.Request)

type tunnelHandler struct {
	upgrader websocket.Upgrader
}

func NewTunnelHandlder(ctx context.Context, wssUpgrader websocket.Upgrader) *tunnelHandler {
	return &tunnelHandler{
		upgrader: wssUpgrader,
	}
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
			if policyRes, _ := authorizer.AuthorizeForward(forwardAction); policyRes {
				//TODO support http2, if is available prefered over WebSocket
				wsConn, err := t.upgrader.Upgrade(res, req, nil)
				if err != nil {
					invalidRequest(res, fmt.Errorf("error during connection upgrade: %v", err))
					return
				}

				webSocketStream(wsConn, dstAddr, netType)
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

func webSocketStream(wsConn *websocket.Conn, dstAddr, netType string) {
	tunnelConn := transport.NewEdgeProxyReadWriter(wsConn)
	dstConn, err := net.Dial(netType, dstAddr)
	if err != nil {
		log.Errorf("Can not connect to %s: %v", dstAddr, err)
		return
	}
	defer dstConn.Close()
	connectionsAccepted.Inc()
	transport.Stream(tunnelConn, dstConn)
}
