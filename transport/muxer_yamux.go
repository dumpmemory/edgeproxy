package transport

import (
	"edgeproxy/server/auth"
	log "github.com/sirupsen/logrus"
	"net"
	"os"
	"time"
)
import "github.com/hashicorp/yamux"

type yamuxMuxer struct {
	yamuxConfig *yamux.Config
}

func NewYamuxMuxer() (*yamuxMuxer, error) {

	yamuxConfig := &yamux.Config{
		AcceptBacklog:          256,
		EnableKeepAlive:        true,
		KeepAliveInterval:      30 * time.Second,
		ConnectionWriteTimeout: 10 * time.Second,
		MaxStreamWindowSize:    256 * 1024,
		StreamCloseTimeout:     5 * time.Minute,
		StreamOpenTimeout:      75 * time.Second,
		LogOutput:              os.Stderr,
	}
	m := &yamuxMuxer{
		yamuxConfig: yamuxConfig,
	}

	return m, nil
}

func (h *yamuxMuxer) ExecuteServerRouter(router *Router, tunnelConn net.Conn, subject string) error {
	session, err := yamux.Server(tunnelConn, h.yamuxConfig)
	if err != nil {
		return err
	}

	for {
		originConn, err := session.Accept()
		if err != nil {
			return err
		}
		go h.acceptConnection(originConn, router, subject)
	}
}

func (h *yamuxMuxer) acceptConnection(originConn net.Conn, router *Router, subject string) {
	defer originConn.Close()
	frame, actionFrame, err := readFrame(originConn)
	if err != nil {
		log.Warnf("error reading frame incoming connection: %v", err)
	}

	switch frame.RouterAction() {
	case ConnectionForwardRouterAction:
		fwFrame, _ := actionFrame.(ForwardFrame)
		forwardAction := auth.NewForwardAction(subject, fwFrame.DstAddr(), fwFrame.NetType().String())
		err = router.ConnectionForward(originConn, forwardAction)
		if err != nil {
			log.Warnf("error on Connection Forward: %v", err)
		}
	}
}
