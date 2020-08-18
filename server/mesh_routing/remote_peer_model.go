package mesh_routing

type RemotePeerViaRPC struct {
	in  chan string
	out chan string
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

func NewRemotePeerViaRPC(in chan string, out chan string) *RemotePeerViaRPC {
	return &RemotePeerViaRPC{in, out}
}
