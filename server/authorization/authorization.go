package authorization

import (
	"crypto/x509"
	"net/http"
)

type Authorizer interface {
	Authorize(w http.ResponseWriter, r *http.Request) (bool, Subject)
}

type Subject interface {
	GetSubject() string
}

type SpiffeSubject struct {
	cert *x509.Certificate
}

func (s SpiffeSubject) GetSubject() string {
	//TODO implement me
	return s.cert.URIs[0].String()
}
