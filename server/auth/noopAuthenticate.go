package auth

import (
	"net/http"
)

var noopAnonymousSubject = &noopSubject{}

type noopAuthAuthorizer struct {
}
type noopSubject struct {
}

func (s noopSubject) GetSubject() string {
	return "anonymous"
}

func NoopAuthorizer() *noopAuthAuthorizer {
	return &noopAuthAuthorizer{}
}

func (*noopAuthAuthorizer) Authenticate(w http.ResponseWriter, r *http.Request) (bool, Subject) {
	return true, noopAnonymousSubject
}

func (*noopAuthAuthorizer) AuthorizeForward(forwardAction ForwardAction) bool {
	return true
}
