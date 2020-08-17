package mesh_routing

import (
	"github.com/google/uuid"
)

type SimplePeer1 struct {
	id string
}

func (th *SimplePeer1) GetID() string {
	return th.id
}

func (th *SimplePeer1) HandleAppearedPeer(id string) {

}
func (th *SimplePeer1) HandleDisappearedPeer(id string) {

}
func (th *SimplePeer1) HandleMessage(id string, data []byte) {

}
func (th *SimplePeer1) RegisterSendMessageHandler(handler func(id string, data []byte)) {

}
func (th *SimplePeer1) HandleTimeTick(ts int64) {

}

func NewSimplePeer1() *SimplePeer1 {
	return &SimplePeer1{uuid.New().String()}
}
