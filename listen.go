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

type listen struct {
	lua.SuperVelaData
	name   string
	url    auxlib.URL
	ne     *kind.Listener
	co     *lua.LState
	banner string
	hook   *pipe.Px
	vsh    *vswitch.Switch
	err    error
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

func (ln *listen) Accept(ctx context.Context, conn net.Conn) error {

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
	ln.hook.CheckMany(L)
	return 0
}

func (ln *listen) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "hook":
		return lua.NewFunction(ln.hookL)
	case "case":
		return ln.vsh.Index(L, "case")
	}
	return lua.LNil
}
