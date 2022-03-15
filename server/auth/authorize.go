package auth

/*type AuthorizationAction uint8

const (
	Forward AuthorizationAction = 0
)*/

type ForwardAction struct {
	Subject         Subject
	DestinationAddr string
	NetType         string
}

func NewForwardAction(subject Subject, destinationAddr, netType string) ForwardAction {
	return ForwardAction{Subject: subject,
		DestinationAddr: destinationAddr, NetType: netType}
}

type Authorize interface {
	AuthorizeForward(forwardAction ForwardAction) bool
}
