package crowd

type MeshActor interface {
	GetID() string
	HandleAppearedPeer(id string)
	HandleDisappearedPeer(id string)
	HandleMessage(id string, data []byte)
	RegisterSendMessageHandler(handler func(id string, data []byte))
	HandleTimeTick(ts int64)
}
