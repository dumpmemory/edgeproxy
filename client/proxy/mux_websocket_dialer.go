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
	rw             sync.RWMutex
	endpoint       *url.URL
	authenticator  clientauth.Authenticator
	ctx            context.Context
	muxSession     *yamux.Session
	ws             recws.RecConn
	forceReconnect chan uint8
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
		ctx:            ctx,
		rw:             sync.RWMutex{},
		endpoint:       endpointUrl,
		authenticator:  authenticator,
		ws:             ws,
		forceReconnect: make(chan uint8, 2),
	}

	err = wssMux.initializeConnection()
	if err != nil {
		return nil, err
	}
	go wssMux.keepAlive()
	go wssMux.monitorConnection()
	return wssMux, nil
}

func (d *muxWebsocketDialer) OpenMuxConnection() (net.Conn, error) {
	conn, err := d.muxSession.Open()
	if err != nil {
		//We wait 5 seconds for reconnection, in case is not capable to get new connection we finally fail
		<-time.After(time.Second * 5)
		return d.muxSession.Open()
	}
	return conn, nil
}

func (d *muxWebsocketDialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	nt, err := transport.NetTypeFromStr(network)
	if err != nil {
		return nil, fmt.Errorf("not Support %s network", network)
	}
	conn, err := d.OpenMuxConnection()
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

func (d *muxWebsocketDialer) initializeConnection() error {
	log.Infof("Connecting to Websocket tunnel endpoint %s", d.endpoint)
	headers := http.Header{}
	headers.Add(transport.HeaderMuxerType, string(transport.YamuxMuxer))
	if d.authenticator != nil {
		d.authenticator.AddAuthenticationHeaders(&headers)
	}
	con, err := stream.NewWebsocketConnFromEndpoint(d.ctx, d.endpoint, headers)
	if err != nil {
		return err
	}

	session, err := yamux.Client(con, nil)
	if err != nil {
		return err
	}

	//Close old connection if exists
	if d.muxSession != nil {
		d.muxSession.Close()
	}

	if d.Conn != nil {
		d.Conn.Close()
	}

	//Assign new
	d.Conn = con
	d.muxSession = session
	log.Infof("Connected to Websocket tunnel %s", d.endpoint)
	return nil
}
func (d *muxWebsocketDialer) monitorConnection() {
	for {
		select {
		case <-d.ctx.Done():
			return
		case <-d.forceReconnect:
			go d.reconnect()
		}
	}
}

func (d *muxWebsocketDialer) reconnect() bool {
	d.rw.Lock()
	defer d.rw.Unlock()
	//Before Reconnect we double check if connection is broken
	_, err := d.muxSession.Ping()
	if err != nil {
		for {
			err := d.initializeConnection()
			if err != nil {
				log.Warnf("Failed on Reconnection: %v", err)
				<-time.After(time.Second * 5)
				continue
			}
			return true
		}
	}
	return false
}
func (d *muxWebsocketDialer) keepAlive() {
	for {
		select {
		case <-d.ctx.Done():
			return
		case <-time.After(time.Second * 5):
			t, err := d.muxSession.Ping()
			if err != nil {
				d.forceReconnect <- 0
			} else {
				log.Debugf("yamux ping: ms %d", t.Milliseconds())
			}

		}
	}

}
