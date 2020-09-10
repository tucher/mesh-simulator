package meshpeer

// NetworkMessage is the lowest level mesh network data package
type NetworkMessage []byte

// NetworkID uniqely identifies a peer in mesh network
type NetworkID string

// NetworkTime provides primitive for timestamps in MeshNetwork
type NetworkTime int64

// MeshAPI defines low-level mesh functionality
type MeshAPI interface {
	GetMyID() NetworkID
	RegisterPeerAppearedHandler(func(id NetworkID))
	RegisterPeerDisappearedHandler(func(id NetworkID))
	RegisterMessageHandler(func(id NetworkID, data NetworkMessage))
	SendMessage(id NetworkID, data NetworkMessage)
	RegisterTimeTickHandler(func(ts NetworkTime))
	SendDebugData(interface{})
}
