package mesh_routing

import (
	crowd "mesh-simulator/crowd_model"
)

type SimplePeer1 struct {
}

func (th *SimplePeer1) HandleAppearedPeer(id crowd.NetworkID) {

}
func (th *SimplePeer1) HandleDisappearedPeer(id crowd.NetworkID) {

}
func (th *SimplePeer1) HandleMessage(id crowd.NetworkID, data crowd.NetworkMessage) {

}
func (th *SimplePeer1) RegisterSendMessageHandler(handler func(id crowd.NetworkID, data crowd.NetworkMessage)) {

}
func (th *SimplePeer1) HandleTimeTick(ts crowd.NetworkTime) {

}

func NewSimplePeer1() *SimplePeer1 {
	return &SimplePeer1{}
}
