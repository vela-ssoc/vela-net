package vnet

import (
	"fmt"
	"github.com/vela-ssoc/vela-kit/kind"
	"github.com/vela-ssoc/vela-kit/lua"
	risk "github.com/vela-ssoc/vela-risk"
	"time"
)

type connection struct {
	conn kind.Conn
	body []byte
	err  error
}

func (cn *connection) String() string                         { return fmt.Sprintf("%p", cn) }
func (cn *connection) Type() lua.LValueType                   { return lua.LTObject }
func (cn *connection) AssertFloat64() (float64, bool)         { return 0, false }
func (cn *connection) AssertString() (string, bool)           { return "", false }
func (cn *connection) AssertFunction() (*lua.LFunction, bool) { return nil, false }
func (cn *connection) Peek() lua.LValue                       { return cn }

func (cn *connection) riskL(L *lua.LState) *risk.Event {
	ev := risk.NewEv()
	ev.Class = risk.CheckClass(L, 1)
	src, port := cn.conn.Source()
	ev.LocalIP = src
	ev.LocalPort = port

	dst, port := cn.conn.Destination()
	ev.RemoteIP = dst
	ev.RemotePort = port
	ev.Payload = string(cn.body)
	ev.Reference = "net.open"
	ev.FromCode = L.CodeVM()
	ev.Time = time.Now()
	ev.Subject = "net.open"
	ev.Alert = true

	return ev
}

func (cn *connection) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "err":
		if cn.err == nil {
			return lua.LNil
		}
		return lua.S2L(cn.err.Error())
	case "raw":
		return lua.S2L(string(cn.body))
	case "risk":
		return cn.riskL(L)

	default:
		return cn.conn.Index(L, key)
	}
}
