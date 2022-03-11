package ipaccess

import (
	"fmt"
	"net"
	"sort"
)

type Policy struct {
	DefaultAllow bool    `json:"defaultAllow",mapstructure:"defaultAllow"`
	Rules        []*Rule `json:"rules",mapstructure:"rules"`
}
type IPNet struct {
	net.IPNet
}
type Rule struct {
	IpNet IPNet `json:"ipnet",mapstructure:"ipnet"`
	Ports []int `json:"ports",mapstructure:"ports"`
	Allow bool  `json:"allow",mapstructure:"allow"`
}

func (i *IPNet) UnmarshalText(text []byte) error {
	netipStr := string(text)
	/*netipStrSlice := strings.Split(netipStr, "/")
	if len(netipStrSlice) == 1 {
		return fmt.Errorf("invalid NetIP Format %s", netipStr)
	}*/
	_, ipnet, err := net.ParseCIDR(netipStr)
	if err != nil {
		return err
	}
	*i = IPNet{*ipnet}
	return nil
}
func (i *IPNet) MarshalText() ([]byte, error) {
	b := []byte(fmt.Sprintf("%s/%s", i.IP.String(), i.Mask.String()))
	return b, nil
}

func NewPolicy(defaultAllow bool, rules []*Rule) (*Policy, error) {
	policy := Policy{
		DefaultAllow: defaultAllow,
		Rules:        rules,
	}

	return &policy, nil
}

func NewRuleByCIDR(prefix *string, ports []int, allow bool) (Rule, error) {
	if prefix == nil || len(*prefix) == 0 {
		return Rule{}, fmt.Errorf("no prefix provided")
	}

	_, ipnet, err := net.ParseCIDR(*prefix)
	if err != nil {
		return Rule{}, fmt.Errorf("unable to parse cidr: %s", *prefix)
	}

	return NewRule(ipnet, ports, allow), nil
}

func NewRule(ipnet *net.IPNet, ports []int, allow bool) Rule {
	rule := Rule{
		IpNet: IPNet{*ipnet},
		Ports: ports,
		Allow: allow,
	}
	return rule
}

func (h *Policy) Allowed(ip net.IP, port int) (bool, *Rule) {
	if len(h.Rules) == 0 {
		return h.DefaultAllow, nil
	}

	for _, rule := range h.Rules {
		if rule.IpNet.Contains(ip) {
			if len(rule.Ports) == 0 {
				return rule.Allow, rule
			} else if pos := sort.SearchInts(rule.Ports, port); pos < len(rule.Ports) && rule.Ports[pos] == port {
				return rule.Allow, rule
			}
		}
	}

	return h.DefaultAllow, nil
}

func (ipr *Rule) String() string {
	return fmt.Sprintf("prefix:%s/port:%s/Allow:%t", ipr.IpNet, ipr.PortsString(), ipr.Allow)
}

func (ipr *Rule) PortsString() string {
	if len(ipr.Ports) > 0 {
		return fmt.Sprint(ipr.Ports)
	}
	return "all"
}
