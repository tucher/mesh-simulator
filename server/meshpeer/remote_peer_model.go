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

	msgSender func(id meshsim.NetworkID, data meshsim.NetworkMessage)
}

// HandleAppearedPeer implements crowd.MeshActor
func (th *RPCPeer) HandleAppearedPeer(id meshsim.NetworkID) {
	type apMsg struct {
		PeerID string
	}
	th.sendRPC("foundPeer", apMsg{string(id)})
}

// HandleDisappearedPeer implements crowd.MeshActor
func (th *RPCPeer) HandleDisappearedPeer(id meshsim.NetworkID) {
	type disapMsg struct {
		PeerID string
	}
	th.sendRPC("lostPeer", disapMsg{string(id)})
}

// HandleMessage implements crowd.MeshActor
func (th *RPCPeer) HandleMessage(id meshsim.NetworkID, data meshsim.NetworkMessage) {
	type rcvMsg struct {
		PeerID string
		Data   string
	}
	th.sendRPC("didReceiveFromPeer", rcvMsg{string(id), string(data)})
}

// RegisterSendMessageHandler implements crowd.MeshActor
func (th *RPCPeer) RegisterSendMessageHandler(handler func(id meshsim.NetworkID, data meshsim.NetworkMessage)) {
	th.msgSender = handler
}

// HandleTimeTick implements crowd.MeshActor
func (th *RPCPeer) HandleTimeTick(ts meshsim.NetworkTime) {
	type tickMsg struct {
		TS int64
	}
	th.sendRPC("tick", tickMsg{int64(ts)})
}
func (th *RPCPeer) sendRPC(cmd string, args interface{}) {
	ba, _ := json.Marshal(args)
	c := jsonCommand{cmd, ba}
	b, _ := json.Marshal(c)
	th.out <- b
}

func (th *RPCPeer) sendAnswer(ok bool, args interface{}, err error) {
	type outputStructure struct {
		Ok    bool
		Error string
		Args  interface{}
	}
	o := outputStructure{}
	o.Ok = ok
	o.Args = args
	o.Error = err.Error()

	b, _ := json.Marshal(o)
	th.out <- b
}

func (th *RPCPeer) run() {
	for msg := range th.in {
		cmd := jsonCommand{}
		if json.Unmarshal(msg, &cmd) != nil {
			th.logger.Printf("parsing error: %v", msg)
		}
		switch cmd.Cmd {
		case "sendToPeer":
			type msgSend struct {
				PeerID string
				Data   string
			}
			args := &msgSend{}
			if json.Unmarshal(cmd.Data, &args) != nil {
				th.logger.Printf("parsing error: %v", msg)
				th.sendAnswer(false, nil, nil)
			} else {
				th.msgSender(meshsim.NetworkID(args.PeerID), meshsim.NetworkMessage(args.Data))
				th.sendAnswer(true, nil, nil)
			}

		default:
			th.logger.Printf("WS MESSAGE: %+v", cmd)
		}
	}
}

// NewRPCPeer returns new RPCPeer
func NewRPCPeer(in chan []byte, out chan []byte, logger *log.Logger) *RPCPeer {
	r := &RPCPeer{in, out, logger, func(id meshsim.NetworkID, data meshsim.NetworkMessage) {
		logger.Println("msg handler not registered", id, data)
	}}
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
