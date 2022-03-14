package authorization

import (
	"net/http"
)

type noopAuthorizer struct {
}

func NoopAuthorizer() *noopAuthorizer {
	return &noopAuthorizer{}
}

func (*noopAuthorizer) Authorize(w http.ResponseWriter, r *http.Request) (bool, Subject) {
	return true, nil
}

func (f *noopAuthorizer) Start() {
}

func (f *noopAuthorizer) Stop() {
}
