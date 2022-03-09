package proxy

import (
	"context"
	"fmt"
	"github.com/elazarl/goproxy"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type HTTPProxy struct {
	ctx   context.Context
	proxy *goproxy.ProxyHttpServer
	srv   *http.Server
}

func NewHttpProxy(ctx context.Context, proxyDialer Dialer, proxyPort int) Proxy {
	proxy := goproxy.NewProxyHttpServer()
	/*	proxy.OnRequest(goproxy.UrlIs("ifconfig.me")).DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		fmt.Printf("asd")
		return req, nil
	})*/
	proxy.ConnectDial = proxyDialer.Dial
	proxy.Tr.DialContext = proxyDialer.DialContext
	proxy.Tr.DialTLSContext = proxyDialer.DialContext
	return &HTTPProxy{
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
}

func (h *HTTPProxy) Stop() {
	log.Infof("Stopping Http Proxy")
	if err := h.srv.Shutdown(h.ctx); err != nil {
		log.Fatal(err)
	}
}
