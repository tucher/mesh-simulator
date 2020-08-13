package crowd

import (
	"log"
	"math"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

type Simulator struct {
	logger *log.Logger
	mtx    *sync.RWMutex

	actors map[string]*Actor

	simTime float64
}

type dataEntry struct {
	ID    string
	TS    int64
	Coord [2]float64
}

type actorInfo struct {
	ID    string
	Coord [2]float64
	Peers []string
}
type Overview struct {
	TS     int64
	Actors map[string]actorInfo
}

func (s *Simulator) AddNPC(count int, placeToAdd [2]float64) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	for i := 0; i < count; i++ {
		rndLat := rand.NormFloat64() * 0.00045
		rndLon := rand.NormFloat64() * 0.00045

		na := Actor{
			ID:           strconv.Itoa(i),
			Coord:        [2]float64{placeToAdd[0] + rndLat, placeToAdd[1] + rndLon},
			currentPeers: []string{},
		}
		for i := 0; i < 3; i++ {
			na.randomAmpl[i] = rand.Float64() * 0.0002
			na.randomFreq[i] = rand.Float64() * 0.01
			na.randomPhase[i] = rand.Float64() * 2 * math.Pi
		}
		na.startCoord[0] = na.Coord[0]
		na.startCoord[1] = na.Coord[1]

		s.actors[na.ID] = &na
	}
}

func (s *Simulator) Clear() {

}

func (s *Simulator) GetOverview() Overview {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	ret := Overview{}
	ret.Actors = make(map[string]actorInfo)
	ret.TS = time.Now().UnixNano() / 1000000
	for _, e := range s.actors {
		ret.Actors[e.ID] = actorInfo{e.ID, e.Coord, s.findPeerActorsIDs(e.ID)}
	}

	return ret
}

func hsin(theta float64) float64 {
	return math.Pow(math.Sin(theta/2), 2)
}

func Distance(latlon1 [2]float64, latlon2 [2]float64) float64 {
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

func (s *Simulator) findPeerActorsIDs(id string) []string {

	ret := []string{}
	for p_id, a := range s.actors {
		if p_id == id {
			continue
		}
		if Distance(s.actors[id].Coord, a.Coord) < 50 {
			ret = append(ret, p_id)
		}
	}

	return ret
}
func (s *Simulator) run() {
	var dt float64 = 0.1

	for {
		time.Sleep(time.Duration(dt*1000) * time.Millisecond)
		s.mtx.Lock()

		for _, a := range s.actors {

			a.Coord[0] = a.startCoord[0]
			a.Coord[1] = a.startCoord[1]

			for i := 0; i < 3; i++ {
				a.Coord[0] += math.Sin(2*math.Pi*a.randomFreq[i]*s.simTime+a.randomPhase[i]) * a.randomAmpl[i]
				a.Coord[1] += math.Cos(2*math.Pi*a.randomFreq[i]*s.simTime+a.randomPhase[i]) * a.randomAmpl[i]
			}

		}
		s.simTime += dt
		s.mtx.Unlock()
	}
}

func New(logger *log.Logger) *Simulator {
	n := Simulator{logger, &sync.RWMutex{}, map[string]*Actor{}, 0}

	go n.run()
	return &n
}
