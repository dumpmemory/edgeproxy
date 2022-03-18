package transport

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const (
	ConnectionForwardRouterAction RouterAction = 0
	TcpNetType                    NetType      = 0
	UdpNetType                    NetType      = 1
	protoVersion                  uint8        = 0
	HeaderMuxerType                            = "X-EDGEPROXY-MUXERTYPE"
	HeaderNetworkType                          = "X-EDGEPROXY-NETWORK"
	HeaderRouterAction                         = "X-EDGEPROXY-ACTION"
	HeaderDstAddress                           = "X-EDGEPROXY-DST"
	versionSize                                = 1
	routerAction                               = 1
	payload                                    = 4
	frameSize                                  = versionSize + routerAction + payload
)

var ErrInvalidFrame = errors.New("invalid Frame")

type Frame []byte
type ForwardFrame []byte

func (h Frame) Version() uint8 {
	return h[0]
}

func (h Frame) RouterAction() RouterAction {
	return RouterAction(h[1])
}
func (h Frame) PayloadSize() uint32 {
	return binary.BigEndian.Uint32(h[2:6])
}

func (h ForwardFrame) NetType() NetType {
	return NetType(h[0])
}
func (h ForwardFrame) DstAddr() string {
	return string(h[1:])
}

func (h ForwardFrame) encode(addr string, netType NetType) {
	h[0] = uint8(netType)
	b := []byte(addr)
	for i := 0; i < len(b); i++ {
		h[i+1] = b[i]
	}
}

func (h Frame) encode(routerAction RouterAction, payloadSize int) {
	h[0] = protoVersion
	h[1] = uint8(routerAction)
	binary.BigEndian.PutUint32(h[2:6], uint32(payloadSize))
}

type RouterAction uint8
type NetType uint8

func RouterActionFromString(routerAction string) (RouterAction, error) {
	switch routerAction {
	case "forward":
		return ConnectionForwardRouterAction, nil
	}
	return 0, fmt.Errorf("router Action %s not available", routerAction)
}
func (r RouterAction) String() string {
	switch r {
	case ConnectionForwardRouterAction:
		return "forward"
	}
	return ""
}
func NetTypeFromStr(netType string) (NetType, error) {
	switch netType {
	case "tcp":
		return TcpNetType, nil
		break
	case "udp":
		return UdpNetType, nil
		break
	}
	return 0, fmt.Errorf("netType %s not available", netType)
}

func (n NetType) String() string {
	switch n {
	case TcpNetType:
		return "tcp"
	case UdpNetType:
		return "udp"
	}
	return ""
}

func readFrame(r io.Reader) (Frame, interface{}, error) {
	frame := Frame(make([]byte, frameSize))
	size, err := io.ReadFull(r, frame)
	if err != nil {
		return nil, nil, err
	}
	if size != frameSize {
		return nil, nil, fmt.Errorf("invalid frame Size %d expected %d: %v", size, frameSize, ErrInvalidFrame)
	}
	if frame.Version() != protoVersion {
		return nil, nil, fmt.Errorf("invalid Frame version %d, expected %d: %v", frame.Version(), protoVersion, ErrInvalidFrame)
	}
	switch frame.RouterAction() {
	case ConnectionForwardRouterAction:
		fwdFrame := ForwardFrame(make([]byte, frame.PayloadSize()))
		fwFrameSize, err := io.ReadFull(r, fwdFrame)
		if err != nil {
			return nil, nil, err
		}
		if fwFrameSize != int(frame.PayloadSize()) {
			return nil, nil, fmt.Errorf("invalid Forward frame Size %d expected %d: %v", size, frameSize, ErrInvalidFrame)
		}

		return frame, fwdFrame, nil
		break

	}
	return nil, nil, fmt.Errorf("invalid Router Action Frame %d", frame.RouterAction())
}
func NewForwardFrame(dstAddr string, netType NetType) (Frame, ForwardFrame) {
	fwdFrame := ForwardFrame(make([]byte, len(dstAddr)+1))
	fwdFrame.encode(dstAddr, netType)

	frame := newFrame(len(fwdFrame))
	return frame, fwdFrame
}

func newFrame(payloadSize int) Frame {
	frame := Frame(make([]byte, frameSize))
	frame.encode(ConnectionForwardRouterAction, payloadSize)
	return frame
}
