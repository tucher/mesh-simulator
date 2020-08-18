package meshpeer

import (
	"log"
	"mesh-simulator/meshsim"
)

// SimplePeer1 provides simplest flood peer strategy
type SimplePeer1 struct {
	logger *log.Logger
}

// HandleAppearedPeer implements crowd.MeshActor
func (th *SimplePeer1) HandleAppearedPeer(id meshsim.NetworkID) {

}

// HandleDisappearedPeer implements crowd.MeshActor
func (th *SimplePeer1) HandleDisappearedPeer(id meshsim.NetworkID) {

}

// HandleMessage implements crowd.MeshActor
func (th *SimplePeer1) HandleMessage(id meshsim.NetworkID, data meshsim.NetworkMessage) {

}

// RegisterSendMessageHandler implements crowd.MeshActor
func (th *SimplePeer1) RegisterSendMessageHandler(handler func(id meshsim.NetworkID, data meshsim.NetworkMessage)) {

}

// HandleTimeTick implements crowd.MeshActor
func (th *SimplePeer1) HandleTimeTick(ts meshsim.NetworkTime) {

}

// NewSimplePeer1 returns new SimplePeer1
func NewSimplePeer1(logger *log.Logger) *SimplePeer1 {
	return &SimplePeer1{logger}
}
