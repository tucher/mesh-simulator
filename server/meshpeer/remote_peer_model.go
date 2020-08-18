package meshpeer

import (
	"encoding/json"
	"log"
	"mesh-simulator/meshsim"
)

// RPCPeer provides RPC controllable mesh network peer
type RPCPeer struct {
	in     chan []byte
	out    chan []byte
	logger *log.Logger
}

// HandleAppearedPeer implements crowd.MeshActor
func (th *RPCPeer) HandleAppearedPeer(id meshsim.NetworkID) {

}

// HandleDisappearedPeer implements crowd.MeshActor
func (th *RPCPeer) HandleDisappearedPeer(id meshsim.NetworkID) {

}

// HandleMessage implements crowd.MeshActor
func (th *RPCPeer) HandleMessage(id meshsim.NetworkID, data meshsim.NetworkMessage) {

}

// RegisterSendMessageHandler implements crowd.MeshActor
func (th *RPCPeer) RegisterSendMessageHandler(handler func(id meshsim.NetworkID, data meshsim.NetworkMessage)) {

}

// HandleTimeTick implements crowd.MeshActor
func (th *RPCPeer) HandleTimeTick(ts meshsim.NetworkTime) {

}

func (th *RPCPeer) run() {
	for msg := range th.in {
		cmd := jsonCommand{}
		if json.Unmarshal(msg, &cmd) != nil {
			th.logger.Printf("parsing error: %v", msg)
		}
		switch cmd.Cmd {
		default:
			th.logger.Printf("WS MESSAGE: %+v", cmd)
		}
	}
}

// NewRPCPeer returns new RPCPeer
func NewRPCPeer(in chan []byte, out chan []byte, logger *log.Logger) *RPCPeer {
	r := &RPCPeer{in, out, logger}
	go r.run()
	return r
}

type jsonCommand struct {
	Cmd  string
	Data json.RawMessage
}

func (th *jsonCommand) SerData(data interface{}) {
	th.Data, _ = json.Marshal(data)
}

func (th *jsonCommand) GetData(obj interface{}) error {
	return json.Unmarshal(th.Data, obj)
}
