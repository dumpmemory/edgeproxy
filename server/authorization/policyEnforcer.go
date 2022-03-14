package authorization

import (
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	fileadapter "github.com/casbin/casbin/v2/persist/file-adapter"
	log "github.com/sirupsen/logrus"
)

type PolicyEnforcer struct {
	Enforcer *casbin.Enforcer
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

func NewPolicyEnforcer(aclPolicyPath string) *PolicyEnforcer {
	a := fileadapter.NewAdapter(aclPolicyPath)

	m, _ := model.NewModelFromString(edgeproxyCasbinModel)
	//e, err := casbin.NewEnforcer("resources/rbac_model.conf", a)
	e, err := casbin.NewEnforcer(m, a)

	if err != nil {
		log.Fatalf("cannot load casbin: %v", err)
	}
	pe := &PolicyEnforcer{
		Enforcer: e,
	}
	return pe
}
