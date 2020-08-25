package meshpeer

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
)

type pkgStateUpdate struct {
	TS   NetworkTime
	Data json.RawMessage
}

type pkgStateUpdateReceivedAck struct {
	TS NetworkTime
}

type pkg struct {
	Type    string
	Content json.RawMessage
}

type peerToPeerSyncer struct {
	lastAttemptTS NetworkTime
	lastTickTime  NetworkTime
	synced        bool
	delay         NetworkTime
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

func (s *peerToPeerSyncer) tick(ts NetworkTime) {
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
	UpdateTS  NetworkTime
}

// SimplePeer1 provides simplest flood peer strategy
type SimplePeer1 struct {
	api     MeshAPI
	logger  *log.Logger
	Label   string
	syncers map[NetworkID]*peerToPeerSyncer

	meshNetworkState map[NetworkID]peerState
	currentTS        NetworkTime

	nextSendTime NetworkTime
}

// HandleAppearedPeer implements crowd.MeshActor
func (th *SimplePeer1) handleAppearedPeer(id NetworkID) {
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
		th.api.SendMessage(id, bt2)
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

func (th *SimplePeer1) handleDisappearedPeer(id NetworkID) {
	delete(th.syncers, id)
}

func (th *SimplePeer1) handleNewIncomingState(sourceID NetworkID, update pkgStateUpdate) {
	newNetworkState := make(map[NetworkID]peerState)
	somethingChanged := false
	if err := json.Unmarshal(update.Data, &newNetworkState); err == nil {
		for id, newPeerState := range newNetworkState {
			if existingPeerState, ok := th.meshNetworkState[id]; !ok {
				somethingChanged = true
				th.meshNetworkState[id] = newPeerState
				th.api.SendDebugData(th.meshNetworkState)
			} else {
				if existingPeerState.UpdateTS < newPeerState.UpdateTS {
					somethingChanged = true
					th.meshNetworkState[id] = newPeerState
					th.api.SendDebugData(th.meshNetworkState)

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

func (th *SimplePeer1) handleMessage(id NetworkID, data NetworkMessage) {
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
		th.api.SendMessage(id, bt2)
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

func (th *SimplePeer1) handleTimeTick(ts NetworkTime) {
	th.currentTS = ts
	for _, s := range th.syncers {
		s.tick(ts)
	}

	if th.currentTS > th.nextSendTime {
		th.nextSendTime = th.currentTS + NetworkTime(3000000+rand.Int63n(5000000))
		th.SetState(PeerUserState{Message: fmt.Sprintf("%v says %v", th.Label, th.currentTS)})
	}
}

// NewSimplePeer1 returns new SimplePeer1
func NewSimplePeer1(label string, logger *log.Logger, api MeshAPI) *SimplePeer1 {
	ret := &SimplePeer1{
		api:              api,
		logger:           logger,
		Label:            label,
		syncers:          make(map[NetworkID]*peerToPeerSyncer),
		meshNetworkState: make(map[NetworkID]peerState),
	}
	api.RegisterMessageHandler(func(id NetworkID, data NetworkMessage) {
		ret.handleMessage(id, data)
	})
	api.RegisterPeerAppearedHandler(func(id NetworkID) {
		ret.handleAppearedPeer(id)
	})
	api.RegisterPeerDisappearedHandler(func(id NetworkID) {
		ret.handleDisappearedPeer(id)
	})
	api.RegisterTimeTickHandler(func(ts NetworkTime) {
		ret.handleTimeTick(ts)
	})

	return ret
}

// SetState updates this peer user data
func (th *SimplePeer1) SetState(p PeerUserState) {
	th.meshNetworkState[th.api.GetMyID()] = peerState{
		UserState: p,
		UpdateTS:  th.currentTS,
	}
	th.api.SendDebugData(th.meshNetworkState)
	serialisedState, err := json.Marshal(th.meshNetworkState)
	if err != nil {
		th.logger.Println(err.Error())
	}

	for _, syncer := range th.syncers {
		syncer.updateData(serialisedState)
	}
}
