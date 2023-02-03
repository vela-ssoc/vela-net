package vnet

import (
	"github.com/vela-ssoc/vela-kit/lua"
	"github.com/vela-ssoc/vela-kit/vela"
	"time"
)

var xEnv vela.Environment

func newLuaNetCat(L *lua.LState) int {
	nc := newNC(L.IsString(1))
	nc.code = L.CodeVM()
	nc.request(L.IsString(2))
	L.Push(nc)
	return 1
}

func newLuaNetOpen(L *lua.LState) int {
	ln := newLuaListen(L)
	proc := L.NewVelaData(ln.name, listenTypeOf)
	if proc.IsNil() {
		goto done
	}

	if e := proc.Data.(*listen).Close(); e != nil {
		L.RaiseError("%s close error %v", ln.name, e)
		return 0
	}

done:
	proc.Set(ln)

	if e := ln.Start(); e != nil {
		ln.V(lua.VTErr, time.Now())
		L.RaiseError("%s start error %v", ln.name, e)
		return 0
	}

	ln.V(lua.VTRun, time.Now())
	L.Push(proc)
	if ln.hook != nil {
		ln.ne.OnAccept(ln.Accept)
	}

	return 1
}

func WithEnv(env vela.Environment) {
	xEnv = env
	nv := lua.NewUserKV()
	nv.Set("ipv4", lua.NewFunction(newLuaIpv4))
	nv.Set("ipv6", lua.NewFunction(newLuaIPv6))
	nv.Set("ip", lua.NewFunction(newLuaIP))
	nv.Set("ping", lua.NewFunction(newLuaPing))
	nv.Set("cat", lua.NewFunction(newLuaNetCat))
	nv.Set("open", lua.NewFunction(newLuaNetOpen))
	xEnv.Global("net", nv)
}
