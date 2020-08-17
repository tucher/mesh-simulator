package crowd

type ActorPhysics struct {
	ID    string
	Coord [2]float64

	currentPeers map[string]struct{}

	startCoord  [2]float64
	randomAmpl  [3]float64
	randomFreq  [3]float64
	randomPhase [3]float64

	actor MeshActor
}
