package proxy

import (
	"context"
	"fmt"
	"github.com/elazarl/goproxy"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
)

type HTTPProxy struct {
	ctx   context.Context
	proxy *goproxy.ProxyHttpServer
	srv   *http.Server
}
type Dialer interface {
	// Dial connects to the given address via the proxy.
	Dial(network, addr string) (c net.Conn, err error)
}

func NewHttpProxy(ctx context.Context, proxyDialer Dialer, proxyPort int) HTTPProxy {
	proxy := goproxy.NewProxyHttpServer()
	proxy.ConnectDial = proxyDialer.Dial
	return HTTPProxy{
		ctx:   ctx,
		proxy: proxy,
		srv: &http.Server{
			Addr:    fmt.Sprintf(":%d", proxyPort),
			Handler: proxy,
		},
	}
}

func (h *HTTPProxy) Start() {
	h.proxy.Verbose = true

	go func() {
		log.Infof("Starting HTTP Proxy at Addr %s", h.srv.Addr)
		err := h.srv.ListenAndServe()
		if err != http.ErrServerClosed {
			log.Fatalf("Http Proxy Client Listen failure: %v", err)
		}
	}()

	go func() {
		<-h.ctx.Done()
		if err := h.srv.Shutdown(h.ctx); err != nil {
			log.Fatal(err)
		}
	}()
}
