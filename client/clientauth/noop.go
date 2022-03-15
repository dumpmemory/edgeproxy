package clientauth

import (
	log "github.com/sirupsen/logrus"
	"net/http"
)

type NoopAuthenticator struct {
}

func (receiver NoopAuthenticator) AddAuthenticationHeaders(headers *http.Header) {
	log.Trace("skipping clientauth")
}
