package vnet

import (
	"context"
	"fmt"
	"github.com/vela-ssoc/vela-kit/auxlib"
	"github.com/vela-ssoc/vela-kit/kind"
	"github.com/vela-ssoc/vela-kit/lua"
	"github.com/vela-ssoc/vela-kit/pipe"
	"io"
	"net"
	"reflect"
	"time"
)

var pipeTypeOf = reflect.TypeOf((*NetPipe)(nil)).String()

type NetPipe struct {
	lua.SuperVelaData
	bind      auxlib.URL
	forward   Forward
	kln       *kind.Listener
	co        *lua.LState
	onConnect *pipe.Chains
	onClose   *pipe.Chains
}

func (np *NetPipe) reset() {
	np.onConnect = nil
	np.onClose = nil

}

func (np *NetPipe) pipe(ctx context.Context, conn net.Conn, pws net.Conn) {
	var (
		readBytes  int64
		writeBytes int64
	)

	ts := time.Now()

	cancel := func(err error) {
		pws.Close()
		conn.Close()
	}

	defer cancel(fmt.Errorf("pipe defer cancel"))

	go func() {
		n, err := io.Copy(pws, conn)
		writeBytes += n
		cancel(err)
		xEnv.Errorf("connection %s  closed: readBytes %d, writeBytes %d, duration %s", conn.RemoteAddr(), readBytes, writeBytes, time.Now().Sub(ts))
	}()

	go func() {
		n, err := io.Copy(conn, pws) //dst: websocket.conn
		readBytes += n
		cancel(err)
		xEnv.Errorf("connection %s  closed: readBytes %d, writeBytes %d, duration %s", conn.RemoteAddr(), readBytes, writeBytes, time.Now().Sub(ts))
	}()
	<-ctx.Done()
}

func (np *NetPipe) OnConnect(ctx context.Context, conn net.Conn) {
	if np.onConnect == nil {
		return
	}

	co := xEnv.Clone(np.co)
	np.onConnect.Do(kind.NewConn(conn), co, func(err error) {
		xEnv.Errorf("net pipe on connect call fail %v", err)
	})
}

func (np *NetPipe) OnClose(ctx context.Context, conn net.Conn) {
	if np.onClose == nil {
		return
	}

	co := xEnv.Clone(np.co)
	defer xEnv.Free(co)

	np.onClose.Do(kind.NewConn(conn), co, func(err error) {
		xEnv.Errorf("net pipe on close call fail %v", err)
	})
}

func (np *NetPipe) Handle(ctx context.Context, conn net.Conn) error {
	np.OnConnect(ctx, conn)
	dst, err := np.forward.Dail(ctx)
	if err != nil {
		xEnv.Errorf("try connect %s  failed: %s", conn.RemoteAddr(), err.Error())
		conn.Close()
		return err
	}
	np.pipe(ctx, conn, dst)
	np.OnClose(ctx, conn)
	return nil
}

func (np *NetPipe) Listen() error {
	ln, err := kind.Listen(xEnv, np.bind)
	if err != nil {
		return err
	}

	np.kln = ln
	xEnv.Errorf("net.pipe listen %s succeed", np.bind.String())

	return np.kln.OnAccept(np.Handle)
}

func (np *NetPipe) Reload() error {
	np.kln.CloseActiveConn()
	return nil
}

func (np *NetPipe) Start() error {
	err := np.Listen()
	if err != nil {
		return err
	}
	np.V(lua.VTRun, time.Now())
	return nil
}

func (np *NetPipe) Close() error {
	if np.kln == nil {
		return nil
	}

	return np.kln.Close()
}

func (np *NetPipe) Name() string {
	return np.bind.String()
}

func (np *NetPipe) Type() string {
	return pipeTypeOf
}
