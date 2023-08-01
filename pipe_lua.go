package vnet

import (
	"github.com/vela-ssoc/vela-kit/auxlib"
	"github.com/vela-ssoc/vela-kit/lua"
	"github.com/vela-ssoc/vela-kit/pipe"
	"time"
)

func (np *NetPipe) OnConnectL(L *lua.LState) int {
	np.onConnect = pipe.NewByLua(L)
	return 0
}

func (np *NetPipe) OnCloseL(L *lua.LState) int {
	np.onClose = pipe.NewByLua(L)
	return 0
}

func (np *NetPipe) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "on_connect":
		return lua.NewFunction(np.OnConnectL)
	case "on_close":
		return lua.NewFunction(np.OnCloseL)
	default:
		return lua.LNil
	}
}

func newNetPipeL(L *lua.LState) int {
	bind := auxlib.CheckURL(L.Get(1), L)
	fwd := CheckForward(L, 2)

	var np *NetPipe

	vda := L.NewVelaData(bind.String(), pipeTypeOf)
	if vda.IsNil() {
		np = &NetPipe{
			bind:    bind,
			forward: fwd,
			co:      L,
		}
		vda.Set(np)
	} else {
		np = vda.Data.(*NetPipe)
		np.forward = fwd
		np.bind = bind
	}

	xEnv.Start(L, np).From(L.CodeVM()).Err(func(err error) {
		np.V(lua.VTErr, time.Now())
		L.RaiseError("start net pipe fail %v", err)
	}).Do()

	L.Push(vda)
	return 1
}
