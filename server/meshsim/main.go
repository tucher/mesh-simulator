package meshsim

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"

	"mesh-simulator/meshpeer"

	"github.com/google/uuid"
)

// Simulator provides the core for mesh network simulator
type Simulator struct {
	logger *log.Logger
	mtx    *sync.RWMutex

	actors map[meshpeer.NetworkID]*actorPhysics

	simTime float64

	timeRatio float64

	totalMsgSendCounter int

	lastStatusTime float64
}

// AddActor adds generic peer to simulation and returns it's id
func (s *Simulator) AddActor(placeToAdd [2]float64, metainfo map[string]interface{}) (meshpeer.MeshAPI, meshpeer.FrontendAPI) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	rndLat := rand.NormFloat64() * 0.00045
	rndLon := rand.NormFloat64() * 0.00045

	newID := meshpeer.NetworkID(uuid.New().String())[0:8]
	na := actorPhysics{
		ID:               newID,
		Coord:            [2]float64{placeToAdd[0] + rndLat, placeToAdd[1] + rndLon},
		currentPeers:     make(map[meshpeer.NetworkID]struct{}),
		outgoingMsgQueue: make(map[meshpeer.NetworkID][]meshpeer.NetworkMessage),
		mtx:              &sync.Mutex{},
		metainfo:         metainfo,
	}
	na.sender = func(id meshpeer.NetworkID, data meshpeer.NetworkMessage) {
		na.mtx.Lock()
		defer na.mtx.Unlock()

		s.totalMsgSendCounter++
		if _, ok := na.outgoingMsgQueue[id]; !ok {
			na.outgoingMsgQueue[id] = []meshpeer.NetworkMessage{}
		}
		na.outgoingMsgQueue[id] = append(na.outgoingMsgQueue[id], data)
	}
	for i := 0; i < 3; i++ {
		na.randomAmpl[i] = rand.Float64() * 0.0002
		na.randomFreq[i] = rand.Float64() * 0.01
		na.randomPhase[i] = rand.Float64() * 2 * math.Pi
	}
	na.startCoord[0] = na.Coord[0]
	na.startCoord[1] = na.Coord[1]

	s.actors[na.ID] = &na

	na.peerAppearedHandler = func(meshpeer.NetworkID) {}
	na.peerDisappearedHandler = func(meshpeer.NetworkID) {}
	na.messageHandler = func(meshpeer.NetworkID, meshpeer.NetworkMessage) {}
	na.timeTickHandler = func(meshpeer.NetworkTime) {}
	na.userDataSetter = func(interface{}) {}
	return &na, &na
}

// RemoveActor removes peer from simulation by it's ID
func (s *Simulator) RemoveActor(id meshpeer.NetworkID) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if _, ok := s.actors[id]; ok {
		delete(s.actors, id)
	}
}

// Overview stores high level information about current simulation state
type Overview struct {
	TS     int64
	Actors map[string]actorInfo
}

type actorInfo struct {
	ID           string
	Coord        [2]float64
	Peers        []string
	Meta         map[string]interface{}
	CurrentState interface{}
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
		ret.Actors[string(e.ID)] = actorInfo{string(e.ID), e.Coord, prs, e.metainfo, e.debugData}
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

func difference(mOld, mNew map[meshpeer.NetworkID]struct{}) (appeared []meshpeer.NetworkID, disappeared []meshpeer.NetworkID) {
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
	ID meshpeer.NetworkID
}
type peerDists []peerDist

func (a peerDists) Len() int           { return len(a) }
func (a peerDists) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a peerDists) Less(i, j int) bool { return a[i].D < a[j].D }

func (s *Simulator) findPeerActorsIDs(id meshpeer.NetworkID, maxDist float64, maxCount int) map[meshpeer.NetworkID]struct{} {

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
	ret := make(map[meshpeer.NetworkID]struct{})

	for i := 0; i < len(dists) && i < maxCount; i++ {
		ret[dists[i].ID] = struct{}{}
	}

	return ret
}
func (s *Simulator) run() {
	var dt float64 = 0.020

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
			a.timeTickHandler(meshpeer.NetworkTime(s.simTime * 1000000))

			type PeerUserState struct {
				Coordinates []float64
				Message     string
			}

			if s.simTime-a.userInterestingEventTime > 10 && s.simTime >= a.nextUserSimulationSentTime {
				a.nextUserSimulationSentTime = s.simTime
				a.userDataSetter(PeerUserState{
					Coordinates: []float64(a.Coord[:]),
					Message:     fmt.Sprintf("It's boring for %vs", int(s.simTime-a.userInterestingEventTime)),
				})
				a.nextUserSimulationSentTime += rand.Float64()*8.0 + 3.0
			}
			for _, app := range appeared {
				a.userInterestingEventTime = s.simTime
				a.peerAppearedHandler(app)
				a.userDataSetter(PeerUserState{
					Coordinates: []float64(a.Coord[:]),
					Message:     fmt.Sprintf("Hi, %v!", app),
				})
			}

			for _, dis := range disappeared {
				a.userInterestingEventTime = s.simTime
				a.peerDisappearedHandler(dis)
				a.userDataSetter(PeerUserState{
					Coordinates: []float64(a.Coord[:]),
					Message:     fmt.Sprintf("Bye, %v!", dis),
				})
			}
			a.currentPeers = newPeers

			for trgID, msgList := range a.outgoingMsgQueue {
				if peer, found := s.actors[trgID]; found {
					if _, found := a.currentPeers[trgID]; found {
						for _, msg := range msgList {
							peer.messageHandler(a.ID, msg)
						}
					}
				}
			}
			a.outgoingMsgQueue = make(map[meshpeer.NetworkID][]meshpeer.NetworkMessage)
		}
		s.simTime += dt

		if s.simTime-s.lastStatusTime > 1 {
			s.lastStatusTime = s.simTime
			s.logger.Println("Total messages sent: ", s.totalMsgSendCounter)
		}
		s.mtx.Unlock()
	}
}

// New creates and start new simulation
func New(logger *log.Logger) *Simulator {
	n := Simulator{
		logger:              logger,
		mtx:                 &sync.RWMutex{},
		actors:              map[meshpeer.NetworkID]*actorPhysics{},
		simTime:             0,
		timeRatio:           1,
		totalMsgSendCounter: 0,
		lastStatusTime:      0,
	}

	return &n
}

// Run starts simulation
func (s *Simulator) Run() {
	go s.run()
}

// SendMessage send message from given peer to given peers. If target peers are empty, sends to all available peers(broadcast)
func (s *Simulator) SendMessage(ID meshpeer.NetworkID, targets []meshpeer.NetworkID, data meshpeer.NetworkMessage) error {
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
				srcPeer.outgoingMsgQueue[pr] = []meshpeer.NetworkMessage{}
			}
			srcPeer.outgoingMsgQueue[pr] = append(srcPeer.outgoingMsgQueue[pr], data)
		}
		return nil
	}
	return fmt.Errorf("Source peer not found")
}
