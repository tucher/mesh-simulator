package meshpeer

import (
	"fmt"
	"log"
	"mesh-simulator/meshsim"
)

// SimplePeer1 provides simplest flood peer strategy
type SimplePeer1 struct {
	logger *log.Logger

	currentPeers map[meshsim.NetworkID]struct{}

	sender       func(id meshsim.NetworkID, data meshsim.NetworkMessage)
	lastSendTime meshsim.NetworkTime
}

// HandleAppearedPeer implements crowd.MeshActor
func (th *SimplePeer1) HandleAppearedPeer(id meshsim.NetworkID) {
	th.logger.Println("APPEARED", id)
	th.currentPeers[id] = struct{}{}
}

// HandleDisappearedPeer implements crowd.MeshActor
func (th *SimplePeer1) HandleDisappearedPeer(id meshsim.NetworkID) {
	th.logger.Println("DISAPPEARED", id)
	delete(th.currentPeers, id)
}

// HandleMessage implements crowd.MeshActor
func (th *SimplePeer1) HandleMessage(id meshsim.NetworkID, data meshsim.NetworkMessage) {
	th.logger.Printf("message from %v received: %v", id, string(data))
}

// RegisterSendMessageHandler implements crowd.MeshActor
func (th *SimplePeer1) RegisterSendMessageHandler(handler func(id meshsim.NetworkID, data meshsim.NetworkMessage)) {
	th.sender = handler
}

// HandleTimeTick implements crowd.MeshActor
func (th *SimplePeer1) HandleTimeTick(ts meshsim.NetworkTime) {
	if ts-th.lastSendTime > 2000000 {
		th.lastSendTime = ts

		for k := range th.currentPeers {
			th.sender(k, []byte(fmt.Sprintf("Hello! time is %v", ts/1000000)))
		}
	}

}

// NewSimplePeer1 returns new SimplePeer1
func NewSimplePeer1(logger *log.Logger) *SimplePeer1 {
	return &SimplePeer1{
		logger,
		make(map[meshsim.NetworkID]struct{}),
		func(id meshsim.NetworkID, data meshsim.NetworkMessage) { logger.Println("Not registered") },
		0,
	}
}
