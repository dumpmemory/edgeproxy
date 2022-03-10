package server

import (
	"context"
	"edgeproxy/server/handlers"
	"fmt"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"sync/atomic"

	"net/http"
)

type httpServer struct {
	ctx         context.Context
	srv         *http.Server
	srvCertPath string
	srvKeyPath  string
	IsReady     *atomic.Value
}

func NewHttpServer(ctx context.Context, httpPort int) httpServer {
	return NewHttpServerWithTLS(ctx, httpPort, "", "")
}

func NewHttpServerWithTLS(ctx context.Context, httpPort int, srvCertPath string, srvKeyPath string) httpServer {
	wsHandler := handlers.NewWebSocketHandler(ctx)
	muxRouter := mux.NewRouter()
	isReady := &atomic.Value{}
	isReady.Store(false)
	muxRouter.HandleFunc("/", wsHandler.SocketHandler)
	muxRouter.HandleFunc("/version", handlers.VersionHandler)
	muxRouter.HandleFunc("/healthz", handlers.Healthz)
	muxRouter.HandleFunc("/readyz", handlers.Readyz(isReady))
	//r.HandleFunc("/stats", StatsHandler)

	return httpServer{
		ctx: ctx,
		srv: &http.Server{
			Addr:    fmt.Sprintf(":%d", httpPort),
			Handler: muxRouter,
		},
		IsReady:     isReady,
		srvCertPath: srvCertPath,
		srvKeyPath:  srvKeyPath,
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
