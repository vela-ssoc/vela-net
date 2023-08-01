package vnet

import (
	"context"
	"github.com/vela-ssoc/vela-kit/auxlib"
	"github.com/vela-ssoc/vela-kit/lua"
	"github.com/vela-ssoc/vela-kit/proxy"
	"net"
	"time"
)

type Forward interface {
	Dail(context.Context) (net.Conn, error)
}

type direct struct {
	url auxlib.URL
}

func (d *direct) Dail(ctx context.Context) (net.Conn, error) {
	timeout := d.url.Int("timeout")
	var dail net.Dialer
	if timeout <= 0 {
		dail = net.Dialer{Timeout: 10 * time.Second}
	} else {
		dail = net.Dialer{Timeout: time.Duration(timeout) * time.Second}
	}
	return dail.Dial(d.url.Scheme(), d.url.Host())
}

func CheckForward(L *lua.LState, idx int) Forward {
	val := L.Get(idx)

	switch val.Type() {
	case lua.LTNil:
		L.RaiseError("invalid forward type got %s", val.Type())
	case lua.LTString:
		return &direct{url: auxlib.CheckURL(val, L)}
	case lua.LTObject:
		return proxy.IsProxy(L, val)

	default:
		L.RaiseError("not found forward")

	}
	return nil
}
