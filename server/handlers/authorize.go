package handlers

import (
	"edgeproxy/server/authorization"
	"edgeproxy/transport"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type httpHandlerFunc func(http.ResponseWriter, *http.Request)

func Authorize(authorizer authorization.Authorizer, acl *authorization.PolicyEnforcer, next httpHandlerFunc) httpHandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		authorized, subject := authorizer.Authorize(res, req)
		if !authorized {
			res.WriteHeader(http.StatusUnauthorized)
			return
		}

		netType := req.Header.Get(transport.HeaderNetworkType)
		dstAddr := req.Header.Get(transport.HeaderDstAddress)
		if subject != nil {
			// run this subject and the requested network action through casbin
			subj := subject.GetSubject()
			if policyRes, _ := acl.Enforcer.Enforce(subj, dstAddr, netType); policyRes {
				next(res, req)
				return
			} else {
				log.Infof("denied %s access to %s/%s", subj, netType, dstAddr)
				res.WriteHeader(http.StatusForbidden)
				return
			}
		} else {
			// this auth method is global, so it doesn't have a subject to evaluate.  let it through
			next(res, req)
		}
	}
}
