package vnet

import (
	"github.com/vela-ssoc/vela-kit/buffer"
	"github.com/vela-ssoc/vela-kit/kind"
	"github.com/vela-ssoc/vela-kit/lua"
	"github.com/vela-ssoc/vela-kit/pipe"
	vswitch "github.com/vela-ssoc/vela-switch"
)

type RevBuffer struct {
	buf  []byte
	rev  *buffer.Byte
	hdp  *pipe.Px
	vsh  *vswitch.Switch
	cnn  kind.Conn
	co   *lua.LState
	over bool
}

var line1 = []byte("\r\n")
var line2 = []byte("\n")

func (r *RevBuffer) append(n int) {
	if r.hdp.Len() == 0 && r.vsh.Len() == 0 {
		return
	}

	if n == 0 {
		return
	}

	chunk := r.buf[:n]
	r.rev.Write(chunk)
}

func (r *RevBuffer) readline(n int) {
	if n == 0 {
		return
	}

	if r.buf[n-1] == '\n' {
		r.call(nil)
	}
}

func (r *RevBuffer) call(err error) {
	defer func() {
		r.over = true
		r.rev.Reset()
	}()

	if r.hdp.Len() == 0 && r.vsh.Len() == 0 {
		return
	}

	conn := &connection{
		conn: r.cnn,
		body: r.buf,
		err:  err,
	}

	r.hdp.Do(conn, r.co, func(err error) {
		//todo
	})

	r.vsh.Do(conn)
}
