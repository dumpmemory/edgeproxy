package clientauth

import "net/http"

type Authenticator interface {
	AddAuthenticationHeaders(headers *http.Header)
}
