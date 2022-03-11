package authorization

import (
	"net/http"
)

type Authorizer interface {
	Authorize(w http.ResponseWriter, r *http.Request) bool
}
