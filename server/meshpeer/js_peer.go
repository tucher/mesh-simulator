package meshpeer

import (
	"encoding/json"
	"log"

	"github.com/dop251/goja"
)

// JSPeer provides environment to run  mesh network peer implemented in JS
type JSPeer struct {
	api       MeshAPI
	jsRuntime *goja.Runtime
	logger    *log.Logger
}

// NewJSPeer returns new RPCPeer
func NewJSPeer(jsCode string, logger *log.Logger, api MeshAPI) (*JSPeer, error) {
	ret := &JSPeer{
		api:       api,
		jsRuntime: goja.New(),
		logger:    logger,
	}
	APIContructor := func(call goja.ConstructorCall) *goja.Object {
		call.This.Set("registerMessageHandler", func(args goja.FunctionCall) goja.Value {
			if len(args.Arguments) != 1 {
				return ret.jsRuntime.ToValue(false)
			}
			f, ok := goja.AssertFunction(args.Arguments[0])
			if !ok {
				return ret.jsRuntime.ToValue(false)
			}
			api.RegisterMessageHandler(func(id NetworkID, data NetworkMessage) {
				if _, err := f(args.This, ret.jsRuntime.ToValue(string(id)), ret.jsRuntime.ToValue(string(data))); err != nil {
					logger.Println(err.Error())
				}
			})
			return ret.jsRuntime.ToValue(true)
		})

		call.This.Set("registerPeerAppearedHandler", func(args goja.FunctionCall) goja.Value {
			if len(args.Arguments) != 1 {
				return ret.jsRuntime.ToValue(false)
			}
			f, ok := goja.AssertFunction(args.Arguments[0])
			if !ok {
				return ret.jsRuntime.ToValue(false)
			}
			api.RegisterPeerAppearedHandler(func(id NetworkID) {
				if _, err := f(args.This, ret.jsRuntime.ToValue(string(id))); err != nil {
					logger.Println(err.Error())
				}
			})
			return ret.jsRuntime.ToValue(true)
		})

		call.This.Set("registerPeerDisappearedHandler", func(args goja.FunctionCall) goja.Value {
			if len(args.Arguments) != 1 {
				return ret.jsRuntime.ToValue(false)
			}
			f, ok := goja.AssertFunction(args.Arguments[0])
			if !ok {
				return ret.jsRuntime.ToValue(false)
			}
			api.RegisterPeerDisappearedHandler(func(id NetworkID) {
				if _, err := f(args.This, ret.jsRuntime.ToValue(string(id))); err != nil {
					logger.Println(err.Error())
				}
			})
			return ret.jsRuntime.ToValue(true)
		})

		call.This.Set("registerTimeTickHandler", func(args goja.FunctionCall) goja.Value {
			if len(args.Arguments) != 1 {
				return ret.jsRuntime.ToValue(false)
			}
			f, ok := goja.AssertFunction(args.Arguments[0])
			if !ok {
				return ret.jsRuntime.ToValue(false)
			}

			api.RegisterTimeTickHandler(func(ts NetworkTime) {
				if _, err := f(args.This, ret.jsRuntime.ToValue(float64(ts))); err != nil {
					logger.Println(err.Error())
				}
			})
			return ret.jsRuntime.ToValue(true)
		})

		call.This.Set("registerUserDataUpdateHandler", func(args goja.FunctionCall) goja.Value {
			if len(args.Arguments) != 1 {
				return ret.jsRuntime.ToValue(false)
			}
			_, ok := goja.AssertFunction(args.Arguments[0])
			if !ok {
				return ret.jsRuntime.ToValue(false)
			}

			return ret.jsRuntime.ToValue(true)
		})

		call.This.Set("getMyID", func(goja.FunctionCall) goja.Value {
			return ret.jsRuntime.ToValue(string(api.GetMyID()))
		})
		call.This.Set("sendMessage", func(args goja.FunctionCall) goja.Value {
			if len(args.Arguments) != 2 {
				panic(ret.jsRuntime.ToValue("id as string and data as string are required"))
			}
			api.SendMessage(
				NetworkID(args.Arguments[0].String()),
				NetworkMessage(args.Arguments[1].String()),
			)
			return goja.Undefined()
		})
		type debugDataStruct struct {
			Message string
		}

		call.This.Set("setDebugMessage", func(args goja.FunctionCall) goja.Value {
			if len(args.Arguments) != 1 {
				panic(ret.jsRuntime.ToValue("serialised JSON string is needed"))
			}
			api.SendDebugData(json.RawMessage(args.Arguments[0].String()))
			return goja.Undefined()
		})

		return nil
	}

	ret.jsRuntime.Set("MeshAPI", APIContructor)
	ret.jsRuntime.Set("log", func(args goja.FunctionCall) goja.Value {
		s := ""
		for _, a := range args.Arguments {
			s += a.String() + "\t"
		}
		ret.logger.Println("JS log message: ", s)
		return goja.Undefined()
	})

	_, err := ret.jsRuntime.RunString(jsCode)
	if err != nil {
		if jserr, ok := err.(*goja.Exception); ok {
			logger.Println("JS ERROR: ", jserr.String())
		} else {
			logger.Println("ERROR: ", jserr.String())
		}
		return nil, err
	}
	return ret, nil
}
