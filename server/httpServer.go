package server

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"net/http"
)

type httpServer struct {
	ctx         context.Context
	srv         *http.Server
	srvCertPath string
	srvKeyPath  string
}

type WebSocketHandler interface {
	// WebSocket Handler
	socketHandler(w http.ResponseWriter, r *http.Request)
}

func NewHttpServer(ctx context.Context, httpPort int) httpServer {
	return NewHttpServerWithTLS(ctx, httpPort, "", "")
}

func NewHttpServerWithTLS(ctx context.Context, httpPort int, srvCertPath string, srvKeyPath string) httpServer {
	wsHandler := NewWebSocketHandler(ctx)

	muxRouter := mux.NewRouter()
	muxRouter.HandleFunc("/", wsHandler.socketHandler)
	//r.HandleFunc("/stats", StatsHandler)

	return httpServer{
		ctx: ctx,
		srv: &http.Server{
			Addr:    fmt.Sprintf(":%d", httpPort),
			Handler: muxRouter,
		},
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
}

func (w *httpServer) Stop() {
	log.Infof("Stopping HTTP Web Server")
	if err := w.srv.Shutdown(w.ctx); err != nil {
		log.Fatal(err)
	}
}
