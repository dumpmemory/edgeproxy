package transport

import (
	"bytes"
	"io"
	"testing"
)
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

func TestReadFrame(t *testing.T) {

	//Invalid  Frame Incomplete Header
	var invalidFrame1 []byte
	invalidFrame1 = append(invalidFrame1, protoVersion, 7, 1, 6)
	_, _, err := readFrame(bytes.NewReader(invalidFrame1))
	assert.ErrorIs(t, err, io.ErrUnexpectedEOF)

	//Invalid Forward Frame Size not match
	var simulatedCon []byte

	f := newFrame(10)
	simulatedCon = f
	simulatedCon = append(simulatedCon, 7, 1, 6, 7)

	_, _, err = readFrame(bytes.NewReader(simulatedCon))
	assert.ErrorIs(t, err, io.ErrUnexpectedEOF)

}
