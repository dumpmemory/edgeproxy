package proxy

import (
	"context"
	"fmt"
	"github.com/armon/go-socks5"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type SocksProxy struct {
	ctx  context.Context
	srv  *socks5.Server
	addr string
}

func NewSocksProxy(ctx context.Context, proxyDialer Dialer, socksPort int) Proxy {
	conf := &socks5.Config{
		Rules: socks5.PermitAll(),
		//Resolver: nil, //Custom Domain Resolution on the cloud side?
		Dial: proxyDialer.DialContext,
	}
	server, err := socks5.New(conf)
	if err != nil {
		log.Fatalf("error when configuring socks5 server: %v", err)
	}
	sockProxy := &SocksProxy{
		ctx:  ctx,
		srv:  server,
		addr: fmt.Sprintf(":%d", socksPort),
	}
	return sockProxy
}

func (s *SocksProxy) Start() {

	go func() {
		log.Infof("Starting Socks Proxy at Addr %s", s.addr)
		err := s.srv.ListenAndServe("tcp", s.addr)
		if err != http.ErrServerClosed {
			log.Fatalf("Socks Proxy Client Listen failure: %v", err)
		}
	}()
}

func (s *SocksProxy) Stop() {
	log.Infof("Stopping Socks Proxy")
	//TODO No shutdown library implemented on socks5 library :(
}
