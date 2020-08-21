package meshpeer

import (
	"encoding/json"
	"log"
)

// RPCPeer provides RPC controllable mesh network peer
type RPCPeer struct {
	api    MeshAPI
	in     chan []byte
	out    chan []byte
	logger *log.Logger
}

// HandleAppearedPeer implements crowd.MeshActor
func (th *RPCPeer) handleAppearedPeer(id NetworkID) {
	type apMsg struct {
		PeerID string
	}
	th.sendRPC("foundPeer", apMsg{string(id)})
}

// HandleDisappearedPeer implements crowd.MeshActor
func (th *RPCPeer) handleDisappearedPeer(id NetworkID) {
	type disapMsg struct {
		PeerID string
	}
	th.sendRPC("lostPeer", disapMsg{string(id)})
}

// HandleMessage implements crowd.MeshActor
func (th *RPCPeer) handleMessage(id NetworkID, data NetworkMessage) {
	type rcvMsg struct {
		PeerID string
		Data   string
	}
	th.sendRPC("didReceiveFromPeer", rcvMsg{string(id), string(data)})
}

func (th *RPCPeer) handleTimeTick(ts NetworkTime) {
	type tickMsg struct {
		TS int64
	}
	th.sendRPC("tick", tickMsg{int64(ts)})
}

func (th *RPCPeer) sendRPC(cmd string, args interface{}) {
	c := rpcCommand{}
	c.Cmd = cmd
	c.SetData(args)
	th.out <- c.Serialise()
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
		cmd := rpcCommand{}
		if cmd.Deserialise(msg) != nil {
			th.logger.Printf("parsing error: %v", msg)
		}
		switch cmd.Cmd {
		case "sendToPeer":
			type msgSend struct {
				PeerID string
				Data   string
			}
			args := &msgSend{}
			if cmd.GetData(&args) != nil {
				th.logger.Printf("parsing error: %v", msg)
				th.sendAnswer(false, nil, nil)
			} else {
				th.api.SendMessage(NetworkID(args.PeerID), NetworkMessage(args.Data))
				th.sendAnswer(true, nil, nil)
			}

		default:
			th.logger.Printf("WS MESSAGE: %+v", cmd)
		}
	}
}

// NewRPCPeer returns new RPCPeer
func NewRPCPeer(in chan []byte, out chan []byte, logger *log.Logger, api MeshAPI) *RPCPeer {
	ret := &RPCPeer{
		api,
		in,
		out,
		logger,
	}
	api.RegisterMessageHandler(func(id NetworkID, data NetworkMessage) {
		ret.handleMessage(id, data)
	})
	api.RegisterPeerAppearedHandler(func(id NetworkID) {
		ret.handleAppearedPeer(id)
	})
	api.RegisterPeerDisappearedHandler(func(id NetworkID) {
		ret.handleDisappearedPeer(id)
	})
	api.RegisterTimeTickHandler(func(ts NetworkTime) {
		ret.handleTimeTick(ts)
	})

	go ret.run()
	return ret
}

type rpcCommand struct {
	Cmd  string
	Args json.RawMessage
}

func (th *rpcCommand) SetData(data interface{}) {
	th.Args, _ = json.Marshal(data)
}

func (th *rpcCommand) GetData(obj interface{}) error {
	return json.Unmarshal(th.Args, obj)
}

func (th *rpcCommand) Serialise() []byte {
	b, _ := json.Marshal(th)
	return b
}

func (th *rpcCommand) Deserialise(d []byte) error {
	return json.Unmarshal(d, th)
}
