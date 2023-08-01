package vnet

import (
	"context"
	"fmt"
	"github.com/vela-ssoc/vela-kit/auxlib"
	"github.com/vela-ssoc/vela-kit/buffer"
	"github.com/vela-ssoc/vela-kit/kind"
	"github.com/vela-ssoc/vela-kit/lua"
	"github.com/vela-ssoc/vela-kit/pipe"
	vswitch "github.com/vela-ssoc/vela-switch"
	"io"
	"net"
	"reflect"
	"time"
)

var listenTypeOf = reflect.TypeOf((*listen)(nil)).String()

/*
	local p = net.open("tcp://127.0.0.1/9090")

	local p = net.pipe("tcp://127.0.0.1:9092" , "tcp://127.0.0.1:9092").start()

	local x = net.pipe("tcp://127.0.0.1/9092" , vela.proxy("127.0.0.1:8080"))

*/

type listen struct {
	lua.SuperVelaData
	name      string
	url       auxlib.URL
	ne        *kind.Listener
	co        *lua.LState
	banner    string
	hook      *pipe.Chains
	onConnect *pipe.Chains
	onClose   *pipe.Chains
	vsh       *vswitch.Switch
	err       error
}

func newLuaListen(L *lua.LState) *listen {

	raw := L.CheckString(1)
	banner := L.IsString(2)

	url, err := auxlib.NewURL(raw)
	if err != nil {
		L.RaiseError("parse %s url error %v", raw, err)
	}

	ln := &listen{
		url:    url,
		hook:   pipe.NewByLua(L, pipe.Seek(2), pipe.Env(xEnv)),
		banner: banner,
		vsh:    vswitch.NewL(L),
		name:   fmt.Sprintf("listen_%s_%d", url.Scheme(), url.Port()),
		co:     xEnv.Clone(L),
	}
	ln.V(lua.VTInit, listenTypeOf, time.Now())

	return ln
}

func (ln *listen) Name() string {
	return ln.name
}

func (ln *listen) Type() string {
	return listenTypeOf
}

func (ln *listen) Start() error {
	ne, err := kind.Listen(xEnv, ln.url)
	if err != nil {
		return err
	}

	ln.ne = ne
	return nil
}

func (ln *listen) Close() error {
	if ln.ne != nil {
		return ln.ne.Close()
	}
	return nil
}

func (ln *listen) Banner(conn net.Conn) {
	if ln.banner == "" {
		return
	}

	conn.Write(lua.S2B(ln.banner))
}

func (ln *listen) OnConnect(ctx context.Context, conn net.Conn) {
	if ln.onConnect == nil {
		return
	}

	co := xEnv.Clone(ln.co)
	defer xEnv.Free(co)

	ln.onConnect.Do(kind.NewConn(conn), co, func(err error) {
		xEnv.Errorf("listen connection pipe call fail %v", err)
	})
}

func (ln *listen) OnClose(ctx context.Context, conn net.Conn) {
	if ln.onClose == nil {
		return
	}

	co := xEnv.Clone(ln.co)
	defer xEnv.Free(co)

	ln.onClose.Do(kind.NewConn(conn), co, func(err error) {
		xEnv.Errorf("listen connection pipe call fail %v", err)
	})
}

func (ln *listen) Accept(ctx context.Context, conn net.Conn) error {

	ln.OnConnect(ctx, conn)

	rev := &RevBuffer{
		rev: buffer.Get(),
		buf: make([]byte, 4096),
		cnn: kind.NewConn(conn),
		hdp: ln.hook,
		vsh: ln.vsh,
		co:  xEnv.Clone(ln.co),
	}

	ln.Banner(conn)
	defer func() {
		_ = conn.Close()
		ln.OnClose(ctx, conn)
	}()

	for {
		select {
		case <-ctx.Done():
			return nil

		default:
			n, err := conn.Read(rev.buf)
			switch err {
			case nil:
				rev.append(n)
				rev.readline(n)
			case io.EOF:
				rev.call(err)
				return err

			default:
				xEnv.Errorf("%s accept error %v", ln.name, err)
				rev.call(err)
				return err
			}
		}
	}
}

func (ln *listen) hookL(L *lua.LState) int {
	sub := pipe.NewByLua(L)
	ln.hook.Merge(sub)
	return 0
}

func (ln *listen) ToLValue() lua.LValue {
	return lua.NewAnyData(ln, lua.Reflect(lua.OFF))
}

func (ln *listen) onConnectL(L *lua.LState) int {
	sub := pipe.NewByLua(L)

	if ln.onConnect == nil {
		ln.onConnect = sub
	} else {
		ln.onConnect.Merge(sub)
	}

	L.Push(ln.ToLValue())
	return 1

}

func (ln *listen) onCloseL(L *lua.LState) int {
	sub := pipe.NewByLua(L)

	if ln.onClose == nil {
		ln.onClose = sub
	} else {
		ln.onClose.Merge(sub)
	}

	L.Push(ln.ToLValue())
	return 1
}

func (ln *listen) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "hook":
		return lua.NewFunction(ln.hookL)
	case "case":
		return ln.vsh.Index(L, "case")
	case "default":
		return ln.vsh.Index(L, "default")
	case "on_connect":
		return lua.NewFunction(ln.onConnectL)
	case "on_close":
		return lua.NewFunction(ln.onCloseL)
	}
	return lua.LNil
}
