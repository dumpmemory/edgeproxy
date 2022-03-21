package server

import (
	"context"
	"crypto/rand"

	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"edgeproxy/server/auth"
	"edgeproxy/server/handlers"
	"encoding/pem"
	"fmt"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/http2"
	"io/ioutil"
	"math/big"
	mrand "math/rand"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"net/http"
)

type httpServer struct {
	ctx          context.Context
	noTLSServer  *http.Server
	tlsServer    *http.Server
	srvCertPath  string
	srvKeyPath   string
	IsReady      *atomic.Value
	authenticate auth.Authenticate
	authorize    auth.Authorize
	http2Srv     *http2.Server
}

func NewHttpServer(ctx context.Context, authorizers auth.Authenticate, authorize auth.Authorize, httpPort, httpsPort int) httpServer {
	return NewHttpServerWithTLS(ctx, authorizers, authorize, httpPort, httpsPort, "", "")
}

func NewHttpServerWithTLS(ctx context.Context, authenticate auth.Authenticate, authorize auth.Authorize, httpPort, httpsPort int, srvCertPath string, srvKeyPath string) httpServer {
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

	noTLS := &http.Server{
		Addr:    fmt.Sprintf(":%d", httpPort),
		Handler: muxRouter,
	}

	tls := &http.Server{
		Addr:    fmt.Sprintf(":%d", httpsPort),
		Handler: muxRouter,
	}

	if srvKeyPath == "" && srvCertPath == "" {
		log.Infof("No Certificate detected, generating random")
		srvKeyPath, srvCertPath = generateRandomCert()
		//TODO should be destroyed when application is stopped

	}

	return httpServer{
		ctx:          ctx,
		noTLSServer:  noTLS,
		tlsServer:    tls,
		authenticate: authenticate,
		authorize:    authorize,
		IsReady:      isReady,
		srvCertPath:  srvCertPath,
		srvKeyPath:   srvKeyPath,
	}
}

func generateRandomCert() (privateKey string, certificate string) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatal("Private key cannot be created.", err.Error())
	}

	tempKey, err := ioutil.TempFile("", "edgeproxy-*.key")
	if err != nil {
		log.Fatal(err)
	}
	defer tempKey.Close()
	tempCrt, err := ioutil.TempFile("", "edgeproxy-*.crt")
	if err != nil {
		log.Fatal(err)
	}
	// Generate a pem block with the private key
	err = pem.Encode(tempKey, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
	if err != nil {
		log.Fatal(err)
	}

	randomNumber := mrand.Intn(999999)
	if err != nil {
		log.Fatal(err)
	}

	tml := x509.Certificate{
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(5, 0, 0),
		SerialNumber: big.NewInt(int64(randomNumber)),
		Subject: pkix.Name{
			CommonName:   "edgeproxy.io",
			Organization: []string{"Edgeproxy Github"},
		},
		BasicConstraintsValid: true,
	}
	cert, err := x509.CreateCertificate(rand.Reader, &tml, &tml, &key.PublicKey, key)
	if err != nil {
		log.Fatal("Certificate cannot be created.", err.Error())
	}

	// Generate a pem block with the certificate
	err = pem.Encode(tempCrt, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert,
	})
	if err != nil {
		log.Fatal(err)
	}
	return tempKey.Name(), tempCrt.Name()
}

func (w *httpServer) Start() {

	go func() {
		log.Infof("Starting HTTP TLS Web Server at Addr %s", w.tlsServer.Addr)
		err := w.tlsServer.ListenAndServeTLS(w.srvCertPath, w.srvKeyPath)
		if err != http.ErrServerClosed {
			log.Fatalf("http server TLS Listen failure: %v", err)
		}
	}()
	go func() {
		log.Infof("Starting HTTP  Web Server at Addr %s", w.noTLSServer.Addr)
		err := w.noTLSServer.ListenAndServe()
		if err != http.ErrServerClosed {
			log.Fatalf("http server  Listen failure: %v", err)
		}
	}()
	w.IsReady.Store(true)
}

func (w *httpServer) Stop() {
	log.Infof("Stopping HTTP Web Server")
	w.noTLSServer.Shutdown(w.ctx)
	w.tlsServer.Shutdown(w.ctx)
}
