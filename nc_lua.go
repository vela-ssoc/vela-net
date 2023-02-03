package vnet

import (
	"github.com/vela-ssoc/vela-kit/kind"
	"github.com/vela-ssoc/vela-kit/lua"
	"strconv"
)

func (nc ncat) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "ok":
		return lua.LBool(nc.ok())
	case "banner":
		return L.NewFunction(nc.banner)
	case "ont":
		return lua.LInt(nc.ont)
	case "cnt":
		return lua.LInt(nc.cnt)

	case "err":
		if nc.err != nil {
			return lua.S2L(nc.err.Error())
		}
		return lua.LNil
	default:
		return lua.LNil
	}
}

func (nc *ncat) banner(L *lua.LState) int {
	port := L.IsInt(1)
	if port != 0 {
		r, ok := nc.info[port]
		if !ok {
			return 0
		}
		L.Push(lua.S2L(r.banner))
	}

	n := len(nc.info)
	if n <= 0 {
		return 0
	}
	buf := kind.NewJsonEncoder()
	buf.Arr("")
	for p, r := range nc.info {
		if r.banner == "" {
			continue
		}
		buf.Tab("")
		buf.KV(strconv.Itoa(p), r.banner)
		buf.End("},")
	}
	buf.End("]")

	L.Push(lua.B2L(buf.Bytes()))
	return 1

}
