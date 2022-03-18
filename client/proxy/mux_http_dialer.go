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
	"io"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type muxHttpDialer struct {
	io.ReadWriteCloser
	rw             sync.RWMutex
	endpoint       *url.URL
	authenticator  clientauth.Authenticator
	ctx            context.Context
	muxSession     *yamux.Session
	ws             recws.RecConn
	forceReconnect chan uint8
	yamuxConfig    *yamux.Config
}

func NewMuxHTTPDialer(ctx context.Context, endpoint string, authenticator clientauth.Authenticator) (*muxHttpDialer, error) {
	endpointUrl, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	ws := recws.RecConn{
		KeepAliveTimeout: 10 * time.Second,
	}

	yamuxConfig := &yamux.Config{
		AcceptBacklog:          256,
		EnableKeepAlive:        true,
		KeepAliveInterval:      30 * time.Second,
		ConnectionWriteTimeout: 10 * time.Second,
		MaxStreamWindowSize:    1024 * 1024,
		StreamCloseTimeout:     5 * time.Minute,
		StreamOpenTimeout:      75 * time.Second,
		LogOutput:              log.StandardLogger().WriterLevel(log.DebugLevel),
	}

	wssMux := &muxHttpDialer{
		ctx:            ctx,
		rw:             sync.RWMutex{},
		endpoint:       endpointUrl,
		authenticator:  authenticator,
		ws:             ws,
		yamuxConfig:    yamuxConfig,
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

func (d *muxHttpDialer) OpenMuxConnection() (net.Conn, error) {
	conn, err := d.muxSession.Open()
	if err != nil {
		//We wait 5 seconds for reconnection, in case is not capable to get new connection we finally fail
		<-time.After(time.Second * 5)
		return d.muxSession.Open()
	}
	return conn, nil
}

func (d *muxHttpDialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	nt, err := transport.NetTypeFromStr(network)
	if err != nil {
		return nil, fmt.Errorf("not Support %s network", network)
	}
	conn, err := d.OpenMuxConnection()
	if err != nil {
		return nil, err
	}
	go func() {
		<-ctx.Done()
		if conn != nil {
			conn.Close()
		}
	}()

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

func (d *muxHttpDialer) Dial(network string, addr string) (net.Conn, error) {
	return d.DialContext(context.Background(), network, addr)
}

func (d *muxHttpDialer) initializeConnection() error {
	log.Infof("Connecting to tunnel endpoint %s", d.endpoint)
	headers := http.Header{}
	headers.Add(transport.HeaderMuxerType, string(transport.YamuxMuxer))
	if d.authenticator != nil {
		d.authenticator.AddAuthenticationHeaders(&headers)
	}
	var conn io.ReadWriteCloser
	conn, err := stream.NewHttpBiStreamConnFromEndpoint(d.ctx, d.endpoint, headers)
	if err != nil {
		return err
	}

	session, err := yamux.Client(conn, d.yamuxConfig)
	if err != nil {
		return err
	}

	//Close old connection if exists
	if d.muxSession != nil {
		d.muxSession.Close()
	}

	if d.ReadWriteCloser != nil {
		d.ReadWriteCloser.Close()
	}

	//Assign new
	d.ReadWriteCloser = conn
	d.muxSession = session
	log.Infof("Connected to tunnel %s", d.endpoint)
	return nil
}
func (d *muxHttpDialer) monitorConnection() {
	for {
		select {
		case <-d.ctx.Done():
			return
		case <-d.forceReconnect:
			go d.reconnect()
		}
	}
}

func (d *muxHttpDialer) reconnect() bool {
	d.rw.Lock()
	defer d.rw.Unlock()
	//Before Reconnect we double check if connection is broken
	_, err := d.muxSession.Ping()
	if err != nil {
		for {
			log.Warnf("Tunnel connection lost, reconnecting...")
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
func (d *muxHttpDialer) keepAlive() {
	for {
		select {
		case <-d.ctx.Done():
			return
		case <-time.After(time.Second * 5):
			log.Debugf("Yamux Num Streams %d", d.muxSession.NumStreams())
			t, err := d.muxSession.Ping()
			if err != nil {
				d.forceReconnect <- 0
			} else {
				log.Debugf("yamux ping: ms %d", t.Milliseconds())
			}

		}
	}

}
