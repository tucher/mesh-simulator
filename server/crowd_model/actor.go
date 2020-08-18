package crowd

type actorPhysics struct {
	ID    NetworkID
	Coord [2]float64

	currentPeers map[NetworkID]struct{}

	startCoord  [2]float64
	randomAmpl  [3]float64
	randomFreq  [3]float64
	randomPhase [3]float64

	actor MeshActor

	outgoingMsgQueue map[NetworkID][]NetworkMessage
}
