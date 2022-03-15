package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	routerForwardAccepted = promauto.NewCounter(prometheus.CounterOpts{
		Name: "edgeproxy_router_forward_accepted",
		Help: "Accepted forwarding connections",
	})
	routerReadBytes = promauto.NewCounter(prometheus.CounterOpts{
		Name: "edgeproxy_router_read_kilobytes",
		Help: "Read bytes by routerForwardAccepted",
	})
	routerWrittenBytes = promauto.NewCounter(prometheus.CounterOpts{
		Name: "edgeproxy_router_written_kilobytes",
		Help: "Written bytes by routerForwardAccepted",
	})
)

func IncrementRouterForwardAcceptedConnections() {
	routerForwardAccepted.Inc()
}

func IncrementRouterReadBytes(readedBytes int64) {
	routerReadBytes.Add(float64(readedBytes) / 1024)
}
func IncrementRouterWrittenBytes(writtenBytes int64) {
	routerWrittenBytes.Add(float64(writtenBytes) / 1024)
}
