package authorization

import (
	"net/http"
)

type noAuthorizer struct {
}

func NoAuthorizer() *noAuthorizer {
	return &noAuthorizer{}
}

func (*noAuthorizer) Authorize(w http.ResponseWriter, r *http.Request) (bool, Subject) {
	return true, nil
}

func (f *noAuthorizer) Start() {
}

func (f *noAuthorizer) Stop() {
}
