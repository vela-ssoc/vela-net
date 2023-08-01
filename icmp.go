package vnet

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/vela-ssoc/vela-kit/kind"
	"github.com/vela-ssoc/vela-kit/lua"
	"net"
	"time"
)

// ICMP 数据包结构体
type ICMP struct {
	Typ      uint8
	Code     uint8
	Checksum uint16
	ID       uint16
	Seq      uint16
}

type ReplyICMP struct {
	Packet ICMP
	Addr   *net.IPAddr

	Cnt  int
	Time int64
	Err  error
}

func (r ReplyICMP) String() string                         { return lua.B2S(r.Byte()) }
func (r ReplyICMP) Type() lua.LValueType                   { return lua.LTObject }
func (r ReplyICMP) AssertFloat64() (float64, bool)         { return 0, false }
func (r ReplyICMP) AssertString() (string, bool)           { return "", false }
func (r ReplyICMP) AssertFunction() (*lua.LFunction, bool) { return nil, false }
func (r ReplyICMP) Peek() lua.LValue                       { return r }

func (r ReplyICMP) ok() bool {
	if r.Err != nil {
		return false
	}

	return true
}

func (r ReplyICMP) Byte() []byte {
	if r.Err != nil {
		return lua.S2B(r.Err.Error())
	}

	enc := kind.NewJsonEncoder()
	enc.Tab("")
	enc.KV("seq", r.Packet.Seq)
	enc.KV("addr", r.Addr.String())
	enc.KV("cnt", r.Cnt)
	enc.KV("time", r.Time)
	enc.KV("code", r.Packet.Code)
	enc.End("}")

	return enc.Bytes()
}

func (r ReplyICMP) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "ok":
		return lua.LBool(r.ok())
	case "addr":
		return lua.S2L(r.Addr.String())
	case "cnt":
		return lua.LInt(r.Cnt)
	case "time":
		return lua.LInt(r.Time)
	case "seq":
		return lua.LInt(r.Packet.Seq)

	case "code":
		return lua.LInt(r.Packet.Code)

	case "id":
		return lua.LInt(r.Packet.ID)

	case "warp":
		if r.Err != nil {
			return lua.S2L(r.Err.Error())
		}
		return lua.LNil

	default:
		return lua.LNil
	}

}

// CheckSum 校验和计算
func iCheckSum(data []byte) uint16 {
	var (
		sum    uint32
		length = len(data)
		index  int
	)
	for length > 1 {
		sum += uint32(data[index])<<8 + uint32(data[index+1])
		index += 2
		length -= 2
	}
	if length > 0 {
		sum += uint32(data[index])
	}
	sum += (sum >> 16)
	return uint16(^sum)
}

func iEcho(dst string) ReplyICMP {
	//构建发送的ICMP包
	icmp := ICMP{
		Typ:      8,
		Code:     0,
		Checksum: 0, //默认校验和为0，后面计算再写入
		ID:       0,
		Seq:      0,
	}

	reply := ReplyICMP{
		Packet: icmp,
	}

	var raddr, err = net.ResolveIPAddr("ip", dst)
	if err != nil {
		reply.Err = err
		return reply
	}

	reply.Addr = raddr

	var buffer bytes.Buffer
	binary.Write(&buffer, binary.BigEndian, icmp)
	icmp.Checksum = iCheckSum(buffer.Bytes())
	buffer.Reset()
	//开始计算时间
	timeStart := time.Now()

	//与目的ip地址建立连接，第二个参数为空则默认为本地ip，第三个参数为目的ip
	con, err := net.DialIP("ip4:icmp", nil, raddr)
	if err != nil {
		reply.Err = err
		return reply
	}

	//主函数接术后关闭连接
	defer con.Close()
	//构建buffer将要发送的数据存入
	var sendBuffer bytes.Buffer
	binary.Write(&sendBuffer, binary.BigEndian, icmp)
	if _, e := con.Write(sendBuffer.Bytes()); e != nil {
		reply.Err = e
		return reply
	}

	//设置读取超时时间为2s
	con.SetReadDeadline((time.Now().Add(time.Second * 2)))
	//构建接受的比特数组
	rec := make([]byte, 1024)
	//读取连接返回的数据，将数据放入rec中
	recCnt, err := con.Read(rec)
	if err != nil {
		reply.Err = err
		return reply
	}

	//设置结束时间，计算两次时间之差为ping的时间
	timeEnd := time.Now()
	durationTime := timeEnd.Sub(timeStart).Nanoseconds() / 1e6
	reply.Cnt = recCnt
	reply.Time = durationTime
	return reply
}

func newLuaPing(L *lua.LState) int {
	dst := L.IsString(1)
	if dst == "" {
		L.Push(ReplyICMP{Err: fmt.Errorf("invalid addr %v", L.Get(1))})
	} else {
		L.Push(iEcho(dst))
	}
	return 1
}
