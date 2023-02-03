package vnet

import (
	"fmt"
	"github.com/vela-ssoc/vela-kit/lua"
	"net"
)

type addr struct {
	ip    net.IP
	ipv6  bool
	ipv4  bool
	valid bool
}

func (a addr) String() string                         { return fmt.Sprintf("%v", a) }
func (a addr) Type() lua.LValueType                   { return lua.LTObject }
func (a addr) AssertFloat64() (float64, bool)         { return 0, false }
func (a addr) AssertString() (string, bool)           { return "", false }
func (a addr) AssertFunction() (*lua.LFunction, bool) { return nil, false }
func (a addr) Peek() lua.LValue                       { return a }

func (a addr) callback(L *lua.LState) int {
	if !a.valid {
		return 0
	}
	top := L.GetTop()
	err := L.PCall(top-1, 0, nil)
	if err != nil {
		L.RaiseError("%v", err)
	}
	return 0
}

func (a addr) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "ipv4":
		return lua.LBool(a.ipv4)
	case "ipv6":
		return lua.LBool(a.ipv6)
	case "valid":
		return L.NewFunction(a.callback)
	default:
		return lua.LNil
	}
}

func newLuaAddr(ip net.IP, mask int) addr {
	if ip == nil {
		return addr{
			valid: false,
			ipv4:  false,
			ipv6:  false,
		}
	}

	val := addr{ip: ip, valid: true}

	switch mask {
	case 4:
		val.ipv4 = true
	case 6:
		val.ipv6 = true
	}

	return val
}

func ipHelper(L *lua.LState) addr {

	if L.GetTop() == 0 {
		return newLuaAddr(nil, 0)
	}

	data := L.Get(1).String()

	ip := net.ParseIP(data)
	if ip == nil {
		return newLuaAddr(nil, 0)
	}

	for _, ch := range data {
		switch ch {
		case '.':
			return newLuaAddr(ip, net.IPv4len)
		case ':':
			return newLuaAddr(ip, net.IPv6len)
		}
	}

	return newLuaAddr(ip, 0)
}

func newLuaIpv4(L *lua.LState) int {
	L.Push(lua.LBool(ipHelper(L).ipv4))
	return 1
}

func newLuaIPv6(L *lua.LState) int {
	L.Push(lua.LBool(ipHelper(L).ipv6))
	return 1
}

func newLuaIP(L *lua.LState) int {
	L.Push(ipHelper(L))
	return 1
}

//net.ipv4("114.114.114.114").valid(queue.push , row)
