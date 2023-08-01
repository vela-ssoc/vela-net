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
	kv := lua.NewUserKV()
	kv.Set("ipv4", lua.NewFunction(newLuaIpv4))
	kv.Set("ipv6", lua.NewFunction(newLuaIPv6))
	kv.Set("ip", lua.NewFunction(newLuaIP))
	kv.Set("ping", lua.NewFunction(newLuaPing))
	kv.Set("cat", lua.NewFunction(newLuaNetCat))
	kv.Set("open", lua.NewFunction(newLuaNetOpen))
	kv.Set("pipe", lua.NewFunction(newNetPipeL))

	xEnv.Global("net",
		lua.NewExport("lua.net.export",
			lua.WithTable(kv)))
}
