package vnet

import (
	"fmt"
	"github.com/vela-ssoc/vela-kit/audit"
	auxlib2 "github.com/vela-ssoc/vela-kit/auxlib"
	"github.com/vela-ssoc/vela-kit/execpt"
	"github.com/vela-ssoc/vela-kit/lua"
	"net"
	"strconv"
	"time"
)

type ncat struct {
	code string
	url  auxlib2.URL
	info map[int]reply
	err  error
	cnt  int
	ont  int
}

func newNC(raw string) ncat {
	nc := ncat{}
	u, err := auxlib2.NewURL(raw)
	if err != nil {
		nc.err = err
		return nc
	}
	nc.info = make(map[int]reply)
	nc.url = u
	return nc
}

func (nc ncat) String() string                         { return fmt.Sprintf("%p", &nc) }
func (nc ncat) Type() lua.LValueType                   { return lua.LTObject }
func (nc ncat) AssertFloat64() (float64, bool)         { return 0, false }
func (nc ncat) AssertString() (string, bool)           { return "", false }
func (nc ncat) AssertFunction() (*lua.LFunction, bool) { return nil, false }
func (nc ncat) Peek() lua.LValue                       { return nc }

func (nc *ncat) ok() bool {
	if nc.err != nil {
		return false
	}
	return true
}

func (nc *ncat) Deadline() time.Duration {
	timeout := nc.url.Int("timeout")
	if timeout == 0 {
		return 200 * time.Millisecond
	}

	return time.Duration(timeout) * time.Millisecond
}

func (nc *ncat) dail(d net.Dialer, scheme string, host string, port int, data string, buf []byte) reply {
	var r reply
	var err error
	var n int

	conn, err := d.Dial(scheme, host)
	if err != nil {
		r.err = err
		goto done
	}

	defer conn.Close()

	r.rAddr = conn.RemoteAddr().String()
	r.lAddr = conn.LocalAddr().String()
	conn.SetDeadline(time.Now().Add(nc.Deadline()))

	if data != "" {
		_, e := conn.Write(auxlib2.S2B(data))
		if e != nil {
			audit.Debug("net cat write error %v", e).From(nc.code).Put()
			goto done
		}
	}

	n, err = conn.Read(buf)
	if err != nil {
		audit.Debug("net cat read error %v", err).From(nc.code).Put()
		goto done
	}
	r.banner = lua.B2S(buf[:n])
	r.cnt = n

done:
	nc.info[port] = r
	return r
}

func (nc *ncat) request(data string) {
	if nc.err != nil {
		return
	}
	d := net.Dialer{Timeout: nc.Deadline()}

	buf := make([]byte, 1024)
	scheme := nc.url.Scheme()
	hostname := nc.url.Hostname()
	port := nc.url.Port()
	if port != 0 {
		r := nc.dail(d, scheme, nc.url.Host(), port, data, buf)
		nc.err = r.err
		return
	}

	me := execpt.New()
	for _, p := range nc.url.Ports() {
		host := hostname + ":" + strconv.Itoa(p)
		r := nc.dail(d, scheme, host, p, data, buf)
		if r.err != nil {
			me.Try(strconv.Itoa(p), r.err)
		} else {
			nc.ont++
		}
	}
	nc.cnt = len(nc.info)
	nc.err = me.Wrap()
}
