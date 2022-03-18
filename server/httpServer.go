package server

import (
	"context"
	"edgeproxy/server/auth"
	"edgeproxy/server/handlers"
	"fmt"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"sync/atomic"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"net/http"
)

type httpServer struct {
	ctx          context.Context
	http1Srv     *http.Server
	srvCertPath  string
	srvKeyPath   string
	IsReady      *atomic.Value
	authenticate auth.Authenticate
	authorize    auth.Authorize
	http2Srv     *http2.Server
}

func NewHttpServer(ctx context.Context, authorizers auth.Authenticate, authorize auth.Authorize, httpPort int) httpServer {
	return NewHttpServerWithTLS(ctx, authorizers, authorize, httpPort, "", "")
}

func NewHttpServerWithTLS(ctx context.Context, authenticate auth.Authenticate, authorize auth.Authorize, httpPort int, srvCertPath string, srvKeyPath string) httpServer {
	muxRouter := mux.NewRouter()
	isReady := &atomic.Value{}
	//TODO this probably should not be defined here
	tunnelHandler := handlers.NewTunnelHandlder(ctx)
	isReady.Store(false)
	muxRouter.HandleFunc("/", tunnelHandler.TunnelHandler(authenticate, authorize))

	muxRouter.HandleFunc("/version", handlers.VersionHandler)
	muxRouter.HandleFunc("/healthz", handlers.Healthz)
	muxRouter.HandleFunc("/readyz", handlers.Readyz(isReady))
	muxRouter.Handle("/metrics", promhttp.Handler())
	http2Srv := &http2.Server{
		MaxHandlers:          0,
		MaxConcurrentStreams: 100,
		IdleTimeout:          60,
	}
	
	http1Srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", httpPort),
		Handler: h2c.NewHandler(muxRouter, http2Srv),
	}
	http2.ConfigureServer(http1Srv, http2Srv)

	return httpServer{
		ctx:      ctx,
		http1Srv: http1Srv,
		http2Srv: http2Srv,

		authenticate: authenticate,
		authorize:    authorize,
		IsReady:      isReady,
		srvCertPath:  srvCertPath,
		srvKeyPath:   srvKeyPath,
	}
}

func (w *httpServer) Start() {
	var err error
	go func() {
		log.Infof("Starting HTTP Web Server at Addr %s", w.http1Srv.Addr)
		if w.srvCertPath != "" && w.srvKeyPath != "" {
			err = w.http1Srv.ListenAndServeTLS(w.srvCertPath, w.srvKeyPath)
		} else {
			err = w.http1Srv.ListenAndServe()
		}

		if err != http.ErrServerClosed {
			log.Fatalf("http Proxy Client Listen failure: %v", err)
		}
	}()
	//TODO this should be only active if we can establish connection to the backend tunnel.
	w.IsReady.Store(true)
}

func (w *httpServer) Stop() {
	log.Infof("Stopping HTTP Web Server")
	if err := w.http1Srv.Shutdown(w.ctx); err != nil {
		log.Fatal(err)
	}
}
