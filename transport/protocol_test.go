package transport

import "testing"
import "github.com/stretchr/testify/assert"

func TestEncodingFrames(t *testing.T) {
	dstAddr := "ifconfig.me:5050"
	netType, err := NetTypeFromStr("tcp")
	assert.NoError(t, err)
	frame, fwdFrame := NewForwardFrame(dstAddr, netType)
	assert.Equal(t, ConnectionForwardRouterAction, frame.RouterAction())

	assert.Equal(t, dstAddr, fwdFrame.DstAddr())
	assert.Equal(t, netType, fwdFrame.NetType())
}

func TestInvalidRouterAction(t *testing.T) {
	_, err := RouterActionFromString("no_valid_action")
	assert.Error(t, err)
}
