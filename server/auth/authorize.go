package auth

type ForwardAction struct {
	Subject         string
	DestinationAddr string
	NetType         string
}

func NewForwardAction(subject, destinationAddr, netType string) ForwardAction {
	return ForwardAction{Subject: subject,
		DestinationAddr: destinationAddr, NetType: netType}
}

type Authorize interface {
	AuthorizeForward(forwardAction ForwardAction) bool
}
