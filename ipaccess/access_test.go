package ipaccess

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"net"
	"testing"
)

func TestConfigLoading(t *testing.T) {
	configExample := `
defaultAllow: false
rules: 
  - ipnet: 192.168.1.0/24
    allow: true
    ports:
      - 80
      - 443
  - ipnet: 155.155.155.15/20
    allow: false
    ports:
      - 9000
`
	b := []byte(configExample)
	policy := &Policy{}
	err := yaml.Unmarshal(b, policy)
	assert.NoError(t, err, "expected proper parsing")

	if len(policy.Rules) != 2 {
		assert.Equal(t, 2, len(policy.Rules), "expected number of Rules")
	}
	rule1 := policy.Rules[0]
	assert.Equal(t, []int{80, 443}, rule1.Ports)
	assert.Equal(t, "192.168.1.0/24", rule1.IpNet.String())
	rule2 := policy.Rules[1]
	assert.Equal(t, []int{9000}, rule2.Ports, "expected port")
	assert.Equal(t, "155.155.144.0/20", rule2.IpNet.String(), "expected mask addr")

	ip1 := net.ParseIP("192.168.1.100")
	accepted, rule := policy.Allowed(ip1, 443)
	assert.Equal(t, rule.String(), rule1.String(), "expected first defined rule")
	assert.Equal(t, true, accepted, "expected ip:port")

	ip2 := net.ParseIP("192.168.2.100")
	accepted, rule = policy.Allowed(ip2, 443)
	assert.Nil(t, rule)
	assert.Equal(t, false, accepted, "expected ip:port not allowed")

	ip3 := net.ParseIP("155.155.155.14")
	accepted, rule = policy.Allowed(ip3, 9000)
	assert.Equal(t, false, accepted, "expected ip:port not allowed")

}
func TestRuleCreationByCIDR(t *testing.T) {
	var cidr *string
	_, err := NewRuleByCIDR(cidr, []int{80}, false)
	assert.Error(t, err, "expected error as cidr is nil")

	badCidr := "1.1.1.1"
	cidr = &badCidr
	_, err = NewRuleByCIDR(cidr, []int{80}, false)
	assert.Error(t, err, "expected error as the cidr is bad")

	goodCidr := "1.1.1.1/24"
	_, ipnet, _ := net.ParseCIDR("1.1.1.0/24")
	cidr = &goodCidr
	rule, err := NewRuleByCIDR(cidr, []int{80}, false)
	assert.NoError(t, err)
	assert.True(t, ipnet.IP.Equal(rule.IpNet.IP) && bytes.Compare(ipnet.Mask, rule.IpNet.Mask) == 0, "ipnet expected to be %+v, got: %+v", ipnet, rule.IpNet)
}
