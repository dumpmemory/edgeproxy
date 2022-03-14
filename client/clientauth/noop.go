package clientauth

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)


type NoopAuthenticator struct {

}

func (receiver NoopAuthenticator) AddAuthenticationHeaders(headers *http.Header)  {
	log.Trace("skipping auth")
}