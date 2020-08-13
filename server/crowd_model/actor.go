package crowd

type Actor struct {
	ID    string
	Coord [2]float64

	currentPeers []string

	startCoord  [2]float64
	randomAmpl  [3]float64
	randomFreq  [3]float64
	randomPhase [3]float64
}
