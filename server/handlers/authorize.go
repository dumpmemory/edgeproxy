package handlers

import (
	"edgeproxy/server/authorization"
	"net/http"
)

type httpHandlerFunc func(http.ResponseWriter, *http.Request)

func Authorize(authorizer authorization.Authorizer, next httpHandlerFunc) httpHandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if !authorizer.Authorize(res, req) {
			res.WriteHeader(http.StatusUnauthorized)
			return
		}
		next(res, req)
	}
}
