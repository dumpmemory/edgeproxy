package auth

import (
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	fileadapter "github.com/casbin/casbin/v2/persist/file-adapter"
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
)

type policyEnforcer struct {
	Enforcer      *casbin.Enforcer
	aclPolicyPath string
}

const edgeproxyCasbinModel = `[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act, eft

[policy_effect]
e = some(where (p.eft == allow)) && !some(where (p.eft == deny))

[matchers]
m = globMatch(r.sub, p.sub) && globMatch(r.obj, p.obj) && r.act == p.act
`

func NewPolicyEnforcer(aclPolicyPath string) *policyEnforcer {
	adapter := fileadapter.NewAdapter(aclPolicyPath)

	model, _ := model.NewModelFromString(edgeproxyCasbinModel)
	enforcer, err := casbin.NewEnforcer(model, adapter)
	enforcer.SetAdapter(adapter)

	if err != nil {
		log.Fatalf("cannot load casbin: %v", err)
	}
	pe := &policyEnforcer{
		Enforcer:      enforcer,
		aclPolicyPath: aclPolicyPath,
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
					log.Debugf("Enforcer policy file change detected")
					if err = p.Enforcer.LoadPolicy(); err != nil {
						log.Error("Error reloading enforcer policy")
					} else {
						log.Infof("Enforcer Policy Updated")
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

	if err = w.Add(p.aclPolicyPath); err != nil {
		return err
	}
	return nil
}
func (p *policyEnforcer) AuthorizeForward(forwardAction ForwardAction) (bool, Subject) {
	ok, err := p.Enforcer.Enforce(forwardAction.Subject.GetSubject(), forwardAction.DestinationAddr, forwardAction.NetType)
	if err != nil {
		log.Error(err)
		return false, nil
	}
	if !ok {
		return false, nil
	}

	return true, forwardAction.Subject
}
