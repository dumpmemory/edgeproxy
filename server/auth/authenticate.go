package auth

import (
	"net/http"
)

type Authenticate interface {
	Authenticate(w http.ResponseWriter, r *http.Request) (bool, Subject)
}

type Subject interface {
	GetSubject() string
}
