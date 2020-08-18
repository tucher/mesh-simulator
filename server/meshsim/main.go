package meshsim

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Simulator provides the core for mesh network simulator
type Simulator struct {
	logger *log.Logger
	mtx    *sync.RWMutex

	actors map[NetworkID]*actorPhysics

	simTime float64

	timeRatio float64
}

type actorInfo struct {
	ID    string
	Coord [2]float64
	Peers []string
	Meta  map[string]interface{}
}

// Overview stores high level information about current simulation state
type Overview struct {
	TS     int64
	Actors map[string]actorInfo
}

// AddActor adds generic peer to simulation and returns it's id
func (s *Simulator) AddActor(actor MeshActor, placeToAdd [2]float64, metainfo map[string]interface{}) (newID NetworkID) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	rndLat := rand.NormFloat64() * 0.00045
	rndLon := rand.NormFloat64() * 0.00045

	newID = NetworkID(uuid.New().String())
	na := actorPhysics{
		ID:               newID,
		Coord:            [2]float64{placeToAdd[0] + rndLat, placeToAdd[1] + rndLon},
		currentPeers:     make(map[NetworkID]struct{}),
		actor:            actor,
		outgoingMsgQueue: make(map[NetworkID][]NetworkMessage),
		mtx:              &sync.Mutex{},
		metainfo:         metainfo,
	}
	for i := 0; i < 3; i++ {
		na.randomAmpl[i] = rand.Float64() * 0.0002
		na.randomFreq[i] = rand.Float64() * 0.01
		na.randomPhase[i] = rand.Float64() * 2 * math.Pi
	}
	na.startCoord[0] = na.Coord[0]
	na.startCoord[1] = na.Coord[1]

	s.actors[na.ID] = &na
	actor.RegisterSendMessageHandler(func(id NetworkID, data NetworkMessage) {
		na.mtx.Lock()
		defer na.mtx.Unlock()
		if _, ok := na.outgoingMsgQueue[id]; !ok {
			na.outgoingMsgQueue[id] = []NetworkMessage{}
		}
		na.outgoingMsgQueue[id] = append(na.outgoingMsgQueue[id], data)
	})

	return
}

// RemoveActor removes peer from simulation by it's ID
func (s *Simulator) RemoveActor(id NetworkID) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if _, ok := s.actors[id]; ok {
		delete(s.actors, id)
	}
}

// GetOverview return current state overview
func (s *Simulator) GetOverview() Overview {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	ret := Overview{}
	ret.Actors = make(map[string]actorInfo)
	ret.TS = time.Now().UnixNano() / 1000000
	for _, e := range s.actors {
		prs := []string{}
		for p := range e.currentPeers {
			prs = append(prs, string(p))
		}
		ret.Actors[string(e.ID)] = actorInfo{string(e.ID), e.Coord, prs, e.metainfo}
	}

	return ret
}

func hsin(theta float64) float64 {
	return math.Pow(math.Sin(theta/2), 2)
}

func distance(latlon1 [2]float64, latlon2 [2]float64) float64 {
	// convert to radians
	// must cast radius as float to multiply later
	var la1, lo1, la2, lo2, r float64
	la1 = latlon1[0] * math.Pi / 180
	lo1 = latlon1[1] * math.Pi / 180
	la2 = latlon2[0] * math.Pi / 180
	lo2 = latlon2[1] * math.Pi / 180

	r = 6378100 // Earth radius in METERS

	// calculate
	h := hsin(la2-la1) + math.Cos(la1)*math.Cos(la2)*hsin(lo2-lo1)

	return 2 * r * math.Asin(math.Sqrt(h))
}

func difference(mOld, mNew map[NetworkID]struct{}) (appeared []NetworkID, disappeared []NetworkID) {
	for x := range mNew {
		if _, found := mOld[x]; !found {
			appeared = append(appeared, x)
		}
	}

	for x := range mOld {
		if _, found := mNew[x]; !found {
			disappeared = append(disappeared, x)
		}
	}

	return
}

type peerDist struct {
	D  float64
	ID NetworkID
}
type peerDists []peerDist

func (a peerDists) Len() int           { return len(a) }
func (a peerDists) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a peerDists) Less(i, j int) bool { return a[i].D < a[j].D }

func (s *Simulator) findPeerActorsIDs(id NetworkID, maxDist float64, maxCount int) map[NetworkID]struct{} {

	dists := peerDists{}
	for pID, a := range s.actors {
		if pID == id {
			continue
		}
		dist := distance(s.actors[id].Coord, a.Coord)

		if dist < maxDist {
			dists = append(dists, peerDist{dist, pID})
		}
	}
	sort.Sort(dists)
	ret := make(map[NetworkID]struct{})

	for i := 0; i < len(dists) && i < maxCount; i++ {
		ret[dists[i].ID] = struct{}{}
	}

	return ret
}
func (s *Simulator) run() {
	var dt float64 = 0.1

	for {
		time.Sleep(time.Duration(dt*1000.0*s.timeRatio) * time.Millisecond)
		s.mtx.Lock()

		for _, a := range s.actors {

			a.Coord[0] = a.startCoord[0]
			a.Coord[1] = a.startCoord[1]

			for i := 0; i < 3; i++ {
				a.Coord[0] += math.Sin(2*math.Pi*a.randomFreq[i]*s.simTime+a.randomPhase[i]) * a.randomAmpl[i]
				a.Coord[1] += math.Cos(2*math.Pi*a.randomFreq[i]*s.simTime+a.randomPhase[i]) * a.randomAmpl[i]
			}

			newPeers := s.findPeerActorsIDs(a.ID, 50, 5)
			appeared, disappeared := difference(a.currentPeers, newPeers)
			a.actor.HandleTimeTick(NetworkTime(s.simTime * 1000000))

			for _, app := range appeared {
				a.actor.HandleAppearedPeer(app)
			}

			for _, dis := range disappeared {
				a.actor.HandleDisappearedPeer(dis)
			}
			a.currentPeers = newPeers

			for trgID, msgList := range a.outgoingMsgQueue {
				if peer, found := s.actors[trgID]; found {
					if _, found := a.currentPeers[trgID]; found {
						for _, msg := range msgList {
							peer.actor.HandleMessage(a.ID, msg)
						}
					}
				}
			}
			a.outgoingMsgQueue = make(map[NetworkID][]NetworkMessage)
		}
		s.simTime += dt
		s.mtx.Unlock()
	}
}

// New creates and start new simulation
func New(logger *log.Logger) *Simulator {
	n := Simulator{logger, &sync.RWMutex{}, map[NetworkID]*actorPhysics{}, 0, 1}

	go n.run()
	return &n
}

// SendMessage send message from given peer to given peers. If target peers are empty, sends to all available peers(broadcast)
func (s *Simulator) SendMessage(ID NetworkID, targets []NetworkID, data NetworkMessage) error {
	if srcPeer, ok := s.actors[ID]; ok {
		srcPeer.mtx.Lock()
		defer srcPeer.mtx.Unlock()
		if len(targets) == 0 {
			for trg := range srcPeer.currentPeers {
				targets = append(targets, trg)
			}
		}

		for _, pr := range targets {
			if _, ok := srcPeer.outgoingMsgQueue[pr]; !ok {
				srcPeer.outgoingMsgQueue[pr] = []NetworkMessage{}
			}
			srcPeer.outgoingMsgQueue[pr] = append(srcPeer.outgoingMsgQueue[pr], data)
		}
		return nil
	}
	return fmt.Errorf("Source peer not found")
}
