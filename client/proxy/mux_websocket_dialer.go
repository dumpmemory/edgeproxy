package proxy

import (
	"context"
	"edgeproxy/client/clientauth"
	"edgeproxy/stream"
	"edgeproxy/transport"
	"fmt"
	"github.com/hashicorp/yamux"
	"github.com/recws-org/recws"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type muxWebsocketDialer struct {
	net.Conn
	rw            sync.RWMutex
	endpoint      *url.URL
	authenticator clientauth.Authenticator
	ctx           context.Context
	muxSession    *yamux.Session
	ws            recws.RecConn
	muxError      error
}

func NewMuxWebSocketDialer(ctx context.Context, endpoint string, authenticator clientauth.Authenticator) (*muxWebsocketDialer, error) {
	endpointUrl, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	ws := recws.RecConn{
		KeepAliveTimeout: 10 * time.Second,
	}

	wssMux := &muxWebsocketDialer{
		ctx:           ctx,
		rw:            sync.RWMutex{},
		endpoint:      endpointUrl,
		authenticator: authenticator,
		ws:            ws,
	}

	err = wssMux.initializeConnection()
	if err != nil {
		return nil, err
	}

	return wssMux, nil
}

func (d *muxWebsocketDialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	nt, err := transport.NetTypeFromStr(network)
	if err != nil {
		return nil, fmt.Errorf("not Support %s network", network)
	}

	conn, err := d.newMuxConn()
	if err != nil {
		return nil, err
	}

	f, fwd := transport.NewForwardFrame(addr, nt)
	_, err = conn.Write(f)
	if err != nil {
		return nil, fmt.Errorf("error when Writting Proto Frame: %v", err)
	}
	_, err = conn.Write(fwd)
	if err != nil {
		return nil, fmt.Errorf("error when Writting Forward Frame: %v", err)
	}
	return conn, nil
}

func (d *muxWebsocketDialer) Dial(network string, addr string) (net.Conn, error) {
	return d.DialContext(context.Background(), network, addr)
}

func (d *muxWebsocketDialer) muxHandler() error {
	d.rw.Lock()
	defer d.rw.Unlock()
	if d.muxError != nil {
		d.muxSession.Close()
		d.muxSession = nil
		d.Conn.Close()
		err := d.initializeConnection()
		if err != nil {
			return err
		}
	}
	if d.muxSession == nil {
		session, err := yamux.Client(d.Conn, nil)
		if err != nil {
			return err
		}
		d.muxSession = session
		d.muxError = nil
	}

	return nil

}

func (d *muxWebsocketDialer) initializeConnection() error {
	log.Infof("Connecting to Websocket tunnel endpoint %s", d.endpoint)
	headers := http.Header{}
	headers.Add(transport.HeaderMuxerType, string(transport.YamuxMuxer))
	if d.authenticator != nil {
		d.authenticator.AddAuthenticationHeaders(&headers)
	}
	con, err := stream.NewWebsocketConnFromEndpoint(d.ctx, d.endpoint, headers)
	if err != nil {
		log.Errorf("error establishing connection to tunnel %s", d.endpoint)
		return err
	}
	d.Conn = con
	log.Infof("Connected to Websocket tunnel %s", d.endpoint)
	return nil
}

func (d *muxWebsocketDialer) newMuxConn() (net.Conn, error) {
	if err := d.muxHandler(); err != nil {
		return nil, err
	}

	conn, err := d.muxSession.Open()
	if err != nil {
		//Could be we lost connection, terminate old one and request new one.
		d.muxError = err
		return nil, fmt.Errorf("error when openning muxer connection: %v", err)
	}
	return conn, nil
}
