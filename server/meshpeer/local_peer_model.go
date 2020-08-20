package meshpeer

import (
	"encoding/json"
	"fmt"
	"log"
	"mesh-simulator/meshsim"
)

type pkgStateUpdate struct {
	TS   meshsim.NetworkTime
	Data json.RawMessage
}

type pkgStateUpdateReceivedAck struct {
	TS meshsim.NetworkTime
}

type pkg struct {
	Type    string
	Content json.RawMessage
}

type peerToPeerSyncer struct {
	lastAttemptTS meshsim.NetworkTime
	lastTickTime  meshsim.NetworkTime
	synced        bool
	delay         meshsim.NetworkTime
	sender        func(pkg pkgStateUpdate)

	updatePkg pkgStateUpdate
}

func (s *peerToPeerSyncer) updateData(data []byte) {
	s.synced = false
	s.lastAttemptTS = 0
	s.updatePkg.Data = data
	s.updatePkg.TS = s.lastTickTime

	s.tick(s.lastTickTime)
}

func (s *peerToPeerSyncer) tick(ts meshsim.NetworkTime) {
	if !s.synced && ts-s.lastAttemptTS >= s.delay {
		s.lastAttemptTS = ts
		s.sender(s.updatePkg)
	}
	s.lastTickTime = ts
}

func (s *peerToPeerSyncer) handleAck(ackPkg pkgStateUpdateReceivedAck) {
	if s.synced {
		return
	}
	if ackPkg.TS == s.updatePkg.TS {
		s.synced = true
	}
}
func newPeerToPeerSyncer(sender func(data pkgStateUpdate)) *peerToPeerSyncer {
	return &peerToPeerSyncer{
		lastAttemptTS: 0,
		lastTickTime:  0,
		synced:        true,
		delay:         30000,
		sender:        sender,
		updatePkg:     pkgStateUpdate{},
	}
}

// PeerUserState contains user data
type PeerUserState struct {
	Coordinates [2]float64
	Message     string
}

type peerState struct {
	UserState PeerUserState
	UpdateTS  meshsim.NetworkTime
}

// SimplePeer1 provides simplest flood peer strategy
type SimplePeer1 struct {
	logger  *log.Logger
	sender  func(id meshsim.NetworkID, data meshsim.NetworkMessage)
	ID      meshsim.NetworkID
	Label   string
	syncers map[meshsim.NetworkID]*peerToPeerSyncer

	meshNetworkState map[meshsim.NetworkID]peerState
	currentTS        meshsim.NetworkTime
}

// HandleAppearedPeer implements crowd.MeshActor
func (th *SimplePeer1) HandleAppearedPeer(id meshsim.NetworkID) {
	th.syncers[id] = newPeerToPeerSyncer(func(d pkgStateUpdate) {
		bt, err := json.Marshal(d)
		if err != nil {
			th.logger.Println(err.Error())
			return
		}
		p := pkg{Type: "pkgStateUpdate", Content: bt}
		bt2, err := json.Marshal(p)
		if err != nil {
			th.logger.Println(err.Error())
			return
		}
		th.sender(id, bt2)
	})

	if len(th.meshNetworkState) > 0 {
		serialisedState, err := json.Marshal(th.meshNetworkState)
		if err != nil {
			th.logger.Println(err.Error())
			return
		}

		th.syncers[id].updateData(serialisedState)
	}
}

// HandleDisappearedPeer implements crowd.MeshActor
func (th *SimplePeer1) HandleDisappearedPeer(id meshsim.NetworkID) {
	delete(th.syncers, id)
}

func (th *SimplePeer1) handleNewIncomingState(sourceID meshsim.NetworkID, update pkgStateUpdate) {
	newNetworkState := make(map[meshsim.NetworkID]peerState)
	somethingChanged := false
	if err := json.Unmarshal(update.Data, &newNetworkState); err == nil {
		for id, newPeerState := range newNetworkState {
			if existingPeerState, ok := th.meshNetworkState[id]; !ok {
				somethingChanged = true
				th.meshNetworkState[id] = newPeerState
			} else {
				if existingPeerState.UpdateTS < newPeerState.UpdateTS {
					somethingChanged = true
					th.meshNetworkState[id] = newPeerState
				}
			}
		}
	} else {
		th.logger.Println(err.Error())
		return
	}

	if somethingChanged {
		serialisedState, err := json.Marshal(th.meshNetworkState)
		if err != nil {
			th.logger.Println(err.Error())
			return
		}

		for id, syncer := range th.syncers {
			if sourceID == id {
				continue
			}
			syncer.updateData(serialisedState)
		}
	}
}

// HandleMessage implements crowd.MeshActor
func (th *SimplePeer1) HandleMessage(id meshsim.NetworkID, data meshsim.NetworkMessage) {
	inpkg := &pkg{}
	err := json.Unmarshal(data, inpkg)
	if err != nil {
		th.logger.Println(err.Error())
		return
	}

	switch inpkg.Type {
	case "pkgStateUpdate":
		update := pkgStateUpdate{}
		json.Unmarshal(inpkg.Content, &update)
		th.handleNewIncomingState(id, update)

		ack := pkgStateUpdateReceivedAck{}
		ack.TS = update.TS
		ser, _ := json.Marshal(ack)
		p := pkg{Type: "pkgStateUpdateReceivedAck", Content: ser}
		bt2, err := json.Marshal(p)
		if err != nil {
			th.logger.Println(err.Error())
			return
		}
		th.sender(id, bt2)
		break
	case "pkgStateUpdateReceivedAck":
		if p, ok := th.syncers[id]; ok {
			ack := pkgStateUpdateReceivedAck{}
			json.Unmarshal(inpkg.Content, &ack)
			p.handleAck(ack)
		}
		break
	}
}

// RegisterMessageSender implements crowd.MeshActor
func (th *SimplePeer1) RegisterMessageSender(handler func(id meshsim.NetworkID, data meshsim.NetworkMessage)) {
	th.sender = handler
}

var testStateSet bool = false

// HandleTimeTick implements crowd.MeshActor
func (th *SimplePeer1) HandleTimeTick(ts meshsim.NetworkTime) {
	th.currentTS = ts
	for _, s := range th.syncers {
		s.tick(ts)
	}

	if !testStateSet {
		testStateSet = true
		th.SetState(PeerUserState{Message: fmt.Sprintf("FFFuuuuuuu from %v", th.Label)})
	}
}

// DebugData implements crowd.MeshActor
func (th *SimplePeer1) DebugData() interface{} {
	return th.meshNetworkState
}

// NewSimplePeer1 returns new SimplePeer1
func NewSimplePeer1(label string, logger *log.Logger) *SimplePeer1 {
	return &SimplePeer1{
		logger:           logger,
		sender:           func(id meshsim.NetworkID, data meshsim.NetworkMessage) { logger.Println("Not registered") },
		ID:               "",
		Label:            label,
		syncers:          make(map[meshsim.NetworkID]*peerToPeerSyncer),
		meshNetworkState: make(map[meshsim.NetworkID]peerState),
	}
}

// SetState updates this peer user data
func (th *SimplePeer1) SetState(p PeerUserState) {
	th.meshNetworkState[th.ID] = peerState{
		UserState: p,
		UpdateTS:  th.currentTS,
	}

	serialisedState, err := json.Marshal(th.meshNetworkState)
	if err != nil {
		th.logger.Println(err.Error())
	}

	for _, syncer := range th.syncers {
		syncer.updateData(serialisedState)
	}
}
