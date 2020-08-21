package meshsim

import (
	"mesh-simulator/meshpeer"
	"sync"
)

type actorPhysics struct {
	ID    meshpeer.NetworkID
	Coord [2]float64

	currentPeers map[meshpeer.NetworkID]struct{}

	startCoord  [2]float64
	randomAmpl  [3]float64
	randomFreq  [3]float64
	randomPhase [3]float64

	outgoingMsgQueue map[meshpeer.NetworkID][]meshpeer.NetworkMessage

	mtx *sync.Mutex

	metainfo map[string]interface{}

	sender func(id meshpeer.NetworkID, data meshpeer.NetworkMessage)

	peerAppearedHandler    func(id meshpeer.NetworkID)
	peerDisappearedHandler func(id meshpeer.NetworkID)
	messageHandler         func(id meshpeer.NetworkID, data meshpeer.NetworkMessage)
	timeTickHandler        func(ts meshpeer.NetworkTime)

	debugData interface{}
}

func (th *actorPhysics) GetMyID() meshpeer.NetworkID {
	return th.ID
}
func (th *actorPhysics) RegisterPeerAppearedHandler(h func(id meshpeer.NetworkID)) {
	th.peerAppearedHandler = h
}
func (th *actorPhysics) RegisterPeerDisappearedHandler(h func(id meshpeer.NetworkID)) {
	th.peerDisappearedHandler = h
}
func (th *actorPhysics) RegisterMessageHandler(h func(id meshpeer.NetworkID, data meshpeer.NetworkMessage)) {
	th.messageHandler = h
}
func (th *actorPhysics) SendMessage(id meshpeer.NetworkID, data meshpeer.NetworkMessage) {
	th.sender(id, data)
}
func (th *actorPhysics) RegisterTimeTickHandler(h func(ts meshpeer.NetworkTime)) {
	th.timeTickHandler = h
}
func (th *actorPhysics) SendDebugData(d interface{}) {
	th.debugData = d
}
