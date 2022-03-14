package server

import (
	"context"
	"edgeproxy/server/auth"
	"edgeproxy/server/handlers"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"sync/atomic"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"net/http"
)

type httpServer struct {
	ctx          context.Context
	srv          *http.Server
	srvCertPath  string
	srvKeyPath   string
	IsReady      *atomic.Value
	authenticate auth.Authenticate
	authorize    auth.Authorize
}

func NewHttpServer(ctx context.Context, authorizers auth.Authenticate, authorize auth.Authorize, httpPort int) httpServer {
	return NewHttpServerWithTLS(ctx, authorizers, authorize, httpPort, "", "")
}

func NewHttpServerWithTLS(ctx context.Context, authenticate auth.Authenticate, authorize auth.Authorize, httpPort int, srvCertPath string, srvKeyPath string) httpServer {
	muxRouter := mux.NewRouter()
	isReady := &atomic.Value{}
	//TODO this probably should not be defined here
	tunnelHandler := handlers.NewTunnelHandlder(ctx, websocket.Upgrader{})
	isReady.Store(false)
	muxRouter.HandleFunc("/", tunnelHandler.TunnelHandler(authenticate, authorize))

	muxRouter.HandleFunc("/version", handlers.VersionHandler)
	muxRouter.HandleFunc("/healthz", handlers.Healthz)
	muxRouter.HandleFunc("/readyz", handlers.Readyz(isReady))
	muxRouter.Handle("/metrics", promhttp.Handler())

	return httpServer{
		ctx: ctx,
		srv: &http.Server{
			Addr:    fmt.Sprintf(":%d", httpPort),
			Handler: muxRouter,
		},

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
		log.Infof("Starting HTTP Web Server at Addr %s", w.srv.Addr)
		if w.srvCertPath != "" && w.srvKeyPath != "" {
			err = w.srv.ListenAndServeTLS(w.srvCertPath, w.srvKeyPath)
		} else {
			err = w.srv.ListenAndServe()
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
	if err := w.srv.Shutdown(w.ctx); err != nil {
		log.Fatal(err)
	}
}
