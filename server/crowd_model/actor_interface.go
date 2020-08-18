package crowd

// NetworkMessage is the lowest level mesh network data package
type NetworkMessage []byte

// NetworkID uniqely identifies a peer in mesh network
type NetworkID string

// NetworkTime provides primitive for timestamps in MeshNetwork
type NetworkTime int64

// MeshActor is abstract peer in Mesh Network which could be handled by this simulator
type MeshActor interface {
	HandleAppearedPeer(id NetworkID)
	HandleDisappearedPeer(id NetworkID)
	HandleMessage(id NetworkID, data NetworkMessage)
	RegisterSendMessageHandler(handler func(id NetworkID, data NetworkMessage))
	HandleTimeTick(ts NetworkTime)
}
