package authorization

import (
	"context"
	"edgeproxy/ipaccess"
	"edgeproxy/transport"
	"github.com/fsnotify/fsnotify"
	"net"
	"net/http"
	"strconv"
	"strings"
)

type fileAuthorizer struct {
	watcher  *fsnotify.Watcher
	ctx      context.Context
	ipPolicy *ipaccess.Policy
}

func NewFileAuthorizer(ctx context.Context, ipPolicy *ipaccess.Policy) *fileAuthorizer {
	authorizer := &fileAuthorizer{
		ctx:      ctx,
		ipPolicy: ipPolicy,
	}
	return authorizer
}

func (f *fileAuthorizer) Authorize(w http.ResponseWriter, r *http.Request) (bool, Subject) {
	var port int
	var err error
	netType := r.Header.Get(transport.HeaderNetworkType)
	dstAddr := r.Header.Get(transport.HeaderDstAddress)
	if netType == "" {
		return false, nil
	}
	if dstAddr == "" {
		return false, nil
	}
	dstPortAddr := strings.Split(dstAddr, ":")
	dstAddr = dstPortAddr[0]

	if len(dstPortAddr) == 1 {
		port = 80
	} else {
		port, err = strconv.Atoi(dstPortAddr[1])
		if err != nil {
			return false, nil
		}
	}
	ip := net.ParseIP(dstAddr)
	allowed, _ := f.ipPolicy.Allowed(ip, port)
	return allowed, nil
}
