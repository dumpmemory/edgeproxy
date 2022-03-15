package auth

import (
	"edgeproxy/config"
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	fileadapter "github.com/casbin/casbin/v2/persist/file-adapter"
	"github.com/casbin/casbin/v2/util"
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"net"
	"strings"
)

type policyEnforcer struct {
	IpEnforcer          *casbin.Enforcer
	DomainEnforcer      *casbin.Enforcer
	ipAclPolicyPath     string
	domainAclPolicyPath string
}

const edgeproxyIpFilteringCasbinModel = `[request_definition]
r = sub, ip, port, proto

[policy_definition]
p = sub, ip, port, proto, eft

[policy_effect]
e = some(where (p.eft == allow)) && !some(where (p.eft == deny))

[role_definition]
g = _, _

[matchers]
m = globMatch(r.sub, p.sub) && ipMatch(r.ip, p.ip) && globMatch(r.port, p.port) && r.proto == p.proto
`

const edgeproxyDomainFilteringCasbinModel = `[request_definition]
r = sub, domain, port, proto

[policy_definition]
p = sub, domain, port, proto, eft

[policy_effect]
e = some(where (p.eft == allow)) && !some(where (p.eft == deny))

[role_definition]
g = _, _

[matchers]
m = g(r.sub, p.sub) && globMatch(r.domain, p.domain) && globMatch(r.port, p.port) && r.proto == p.proto
`

func NewPolicyEnforcer(aclPolicyPath config.AclCollection) *policyEnforcer {

	// IP ACL is mandatory, even if it's just 0.0.0.0/0 on all ports to everyone
	ipAdapter := fileadapter.NewAdapter(aclPolicyPath.IpPath)
	ipModel, _ := model.NewModelFromString(edgeproxyIpFilteringCasbinModel)

	ipEnforcer, err := casbin.NewEnforcer(ipModel, ipAdapter)
	if err != nil {
		log.Fatalf("cannot load casbin ipModel: %v", err)
	}
	ipEnforcer.SetAdapter(ipAdapter)
	var domainEnforcer *casbin.Enforcer
	var domainEnforcerErr error

	// add a second enforcer for domain-name matching that can use globs
	if aclPolicyPath.DomainPath != "" {
		domainAdapter := fileadapter.NewAdapter(aclPolicyPath.DomainPath)
		domainModel, _ := model.NewModelFromString(edgeproxyDomainFilteringCasbinModel)

		domainEnforcer, domainEnforcerErr = casbin.NewEnforcer(domainModel, domainAdapter)
		if domainEnforcerErr != nil {
			log.Fatalf("cannot load casbin domainModel: %v", domainEnforcerErr)
		}
		domainEnforcer.SetAdapter(domainAdapter)
		// https://casbin.org/docs/en/rbac#use-pattern-matching-in-rbac
		domainEnforcer.AddNamedMatchingFunc("g", "", util.KeyMatch)
		domainEnforcer.BuildRoleLinks()
	}

	pe := &policyEnforcer{
		IpEnforcer:          ipEnforcer,
		DomainEnforcer:      domainEnforcer,
		ipAclPolicyPath:     aclPolicyPath.IpPath,
		domainAclPolicyPath: aclPolicyPath.DomainPath,
	}
	go pe.watchForPolicyChanges()
	return pe
}
func (p *policyEnforcer) watchForPolicyChanges() error {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	go func() {
		defer w.Close()
		for {
			select {
			case event, ok := <-w.Events:
				if !ok {
					return
				}
				log.Debugf("event: %v", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Debugf("Policy file change detected, reloading")
					if err = p.IpEnforcer.LoadPolicy(); err != nil {
						log.Error("Error reloading IpEnforcer policy")
					} else {
						log.Infof("IpEnforcer Policy Updated")
					}
					if p.DomainEnforcer != nil {
						if err = p.DomainEnforcer.LoadPolicy(); err != nil {
							log.Error("Error reloading DomainEnforcer policy")
						} else {
							log.Infof("DomainEnforcer Policy Updated")
						}
						p.DomainEnforcer.BuildRoleLinks()
					}
				}
			case err, ok := <-w.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	if err = w.Add(p.ipAclPolicyPath); err != nil {
		return err
	}
	if p.domainAclPolicyPath != "" {
		if err = w.Add(p.domainAclPolicyPath); err != nil {
			return err
		}
	}
	return nil
}
func (p *policyEnforcer) AuthorizeForward(forwardAction ForwardAction) bool {
	splitAddress := strings.Split(forwardAction.DestinationAddr, ":")
	host := splitAddress[0]

	var authorized bool
	var authErr error

	// determine if IP or hostname
	addr := net.ParseIP(host)
	if addr != nil {
		authorized, authErr = p.IpEnforcer.Enforce(forwardAction.Subject.GetSubject(), host, splitAddress[1], forwardAction.NetType)
	} else {
		authorized, authErr = p.DomainEnforcer.Enforce(forwardAction.Subject.GetSubject(), host, splitAddress[1], forwardAction.NetType)
		if authErr != nil && authorized {
			// TODO: resolve the IP address, then additionally check it against the IpEnforcer
		}
	}

	if authErr != nil {
		log.Error(authErr)
		return false
	}
	if !authorized {
		return false
	}

	return true
}
