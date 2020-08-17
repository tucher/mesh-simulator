package mesh_routing

import (
	"github.com/google/uuid"
)

type RemotePeerViaRPC struct {
	id  string
	in  chan string
	out chan string
}

func (th *RemotePeerViaRPC) GetID() string {
	return th.id
}

func (th *RemotePeerViaRPC) HandleAppearedPeer(id string) {

}
func (th *RemotePeerViaRPC) HandleDisappearedPeer(id string) {

}
func (th *RemotePeerViaRPC) HandleMessage(id string, data []byte) {

}
func (th *RemotePeerViaRPC) RegisterSendMessageHandler(handler func(id string, data []byte)) {

}
func (th *RemotePeerViaRPC) HandleTimeTick(ts int64) {

}

func NewRemotePeerViaRPC(rpcConnID string, in chan string, out chan string) *RemotePeerViaRPC {
	return &RemotePeerViaRPC{uuid.New().String(), in, out}
}
