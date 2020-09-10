package meshpeer

import (
	"encoding/json"
	"log"

	"github.com/dop251/goja"
)

// JSPeer provides environment to run  mesh network peer implemented in JS
type JSPeer struct {
	jsRuntime *goja.Runtime
	logger    *log.Logger
}

// NewJSPeer returns new RPCPeer
func NewJSPeer(jsCode string, logger *log.Logger, meshAPI MeshAPI, frontendAPI FrontendAPI) (*JSPeer, error) {
	ret := &JSPeer{
		jsRuntime: goja.New(),
		logger:    logger,
	}
	meshAPIObj := ret.jsRuntime.NewObject()

	meshAPIObj.Set("registerMessageHandler", func(args goja.FunctionCall) goja.Value {
		if len(args.Arguments) != 1 {
			return ret.jsRuntime.ToValue(false)
		}
		f, ok := goja.AssertFunction(args.Arguments[0])
		if !ok {
			return ret.jsRuntime.ToValue(false)
		}
		meshAPI.RegisterMessageHandler(func(id NetworkID, data NetworkMessage) {
			if _, err := f(args.This, ret.jsRuntime.ToValue(string(id)), ret.jsRuntime.ToValue(string(data))); err != nil {
				logger.Println(err.Error())
			}
		})
		return ret.jsRuntime.ToValue(true)
	})

	meshAPIObj.Set("registerPeerAppearedHandler", func(args goja.FunctionCall) goja.Value {
		if len(args.Arguments) != 1 {
			return ret.jsRuntime.ToValue(false)
		}
		f, ok := goja.AssertFunction(args.Arguments[0])
		if !ok {
			return ret.jsRuntime.ToValue(false)
		}
		meshAPI.RegisterPeerAppearedHandler(func(id NetworkID) {
			if _, err := f(args.This, ret.jsRuntime.ToValue(string(id))); err != nil {
				logger.Println(err.Error())
			}
		})
		return ret.jsRuntime.ToValue(true)
	})

	meshAPIObj.Set("registerPeerDisappearedHandler", func(args goja.FunctionCall) goja.Value {
		if len(args.Arguments) != 1 {
			return ret.jsRuntime.ToValue(false)
		}
		f, ok := goja.AssertFunction(args.Arguments[0])
		if !ok {
			return ret.jsRuntime.ToValue(false)
		}
		meshAPI.RegisterPeerDisappearedHandler(func(id NetworkID) {
			if _, err := f(args.This, ret.jsRuntime.ToValue(string(id))); err != nil {
				logger.Println(err.Error())
			}
		})
		return ret.jsRuntime.ToValue(true)
	})

	meshAPIObj.Set("registerTimeTickHandler", func(args goja.FunctionCall) goja.Value {
		if len(args.Arguments) != 1 {
			return ret.jsRuntime.ToValue(false)
		}
		f, ok := goja.AssertFunction(args.Arguments[0])
		if !ok {
			return ret.jsRuntime.ToValue(false)
		}

		meshAPI.RegisterTimeTickHandler(func(ts NetworkTime) {
			if _, err := f(args.This, ret.jsRuntime.ToValue(float64(ts))); err != nil {
				logger.Println(err.Error())
			}
		})
		return ret.jsRuntime.ToValue(true)
	})

	meshAPIObj.Set("getMyID", func(goja.FunctionCall) goja.Value {
		return ret.jsRuntime.ToValue(string(meshAPI.GetMyID()))
	})
	meshAPIObj.Set("sendMessage", func(args goja.FunctionCall) goja.Value {
		if len(args.Arguments) != 2 {
			panic(ret.jsRuntime.ToValue("id as string and data as string are required"))
		}
		meshAPI.SendMessage(
			NetworkID(args.Arguments[0].String()),
			NetworkMessage(args.Arguments[1].String()),
		)
		return goja.Undefined()
	})

	meshAPIObj.Set("setDebugMessage", func(args goja.FunctionCall) goja.Value {
		if len(args.Arguments) != 1 {
			panic(ret.jsRuntime.ToValue("serialised JSON string is needed"))
		}
		meshAPI.SendDebugData(json.RawMessage(args.Arguments[0].String()))
		return goja.Undefined()
	})
	ret.jsRuntime.Set("meshAPI", meshAPIObj)

	frontendAPIObj := ret.jsRuntime.NewObject()
	frontendAPIObj.Set("registerUserDataUpdateHandler", func(args goja.FunctionCall) goja.Value {
		if len(args.Arguments) != 1 {
			return ret.jsRuntime.ToValue(false)
		}
		f, ok := goja.AssertFunction(args.Arguments[0])
		if !ok {
			return ret.jsRuntime.ToValue(false)
		}

		frontendAPI.RegisterUserDataUpdateHandler(func(d FrontendUserDataType) {
			if _, err := f(args.This, ret.jsRuntime.ToValue(d)); err != nil {
				logger.Println(err.Error())
			}
		})

		return ret.jsRuntime.ToValue(true)
	})

	frontendAPIObj.Set("handleUpdate", func(args goja.FunctionCall) goja.Value {
		if len(args.Arguments) != 1 {
			panic(ret.jsRuntime.ToValue("pass single object with update data"))
		}
		ob := FrontEndUpdateObject{}
		if err := ret.jsRuntime.ExportTo(args.Arguments[0], &ob); err != nil {
			panic(ret.jsRuntime.ToValue(err.Error()))
		}
		frontendAPI.HandleUpdate(ob)
		return goja.Undefined()
	})

	ret.jsRuntime.Set("frontendAPI", frontendAPIObj)

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
