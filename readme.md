# net
> 基础网络库方法

内置方法:

- [net.ipv4(addr)](#ip)
- [net.ipv6(addr)](#ip)
- [net.ip(addr)](#ip)
- [net.ping(addr)](#ping)
- [net.cat(addr)](#cat)
- [net.open(addr)](#open)

## ping
> p = net.ping(addr)

返回值字段:
- ok
- addr
- cnt &emsp;返回的数据
- time
- seq
- code
- id
- warp &emsp;报错

```lua
    local p = net.ping("127.0.0.1")
    print(p.ok)   -- true
    print(p.addr) -- 127.0.0.1
    print(p.cnt)  -- 32
    print(p.time) -- 56
    print(p.code) -- 200
    print(p.id)   -- 123091
    print(p.warp) -- error

```

## ip
> ip = net.ip("127.0.0.1") 获取IP地址对象 <br />
- ipv4 &emsp; 判断是IPv4 <br />
- ipv6 &emsp; 判断是IPv6 <br />

内置字段:
- ip.ipv4
- ip.ipv6
- ip.invalid

```lua
    local v =  net.ipv4("127.0.0.1")
    print(v) --true

    local v = net.ipv6("aa::22")
    print(v) --true

    local v = net.ip("10.0.0.1")
    print(v.ipv4)  --true
    print(v.ipv6)  --false
    print(v.invalid)  --false
```

## cat
> c = net.cat(url , [body]) <br />
> url: 请求标识 tcp://127.0.0.1:8080

返回值字段:
- ok
- banner
- ont
- cnt
- err

```lua
    local c = net.cat("tcp://127.0.0.1:22")
    print(ok)     -- true
    print(banner) -- openssh-xxx-xx
    --todo
```

## open
> fd = net.open(url , banner)  <br />
> url: 请求标识 tcp://127.0.0.1:8080 , banner:显示banner信息 , function: 处理请求[connection](#connection)

内置方法:
- [fd.hook(pipe)](#)
- [fd.case(cnd)](#) &emsp;匹配条件[switch](/switch#case)

```lua
    local fd = net.open("tcp://127.0.0.1:9999" , "HTTP2.0")
    fd.hook(function(conn)
        --todo        
    end)
    fd.case("src == 127.0.0.1").pipe(print)
```


## connection
> 链接套接字信息
- [push(v)](#)
- [pushf(format , v...)](#)
- err &emsp;报错
- raw &emsp;请求数据
- src
- dst
- src_port
- dst_port
- source &emsp;127.0.0.1:9120
- destination &emsp;127.0.0.1:80

```lua
    fd = net.open("tcp://127.0.0.1:9029" , "ssh2.0")
    fd.hook(function(conn) 
        print(conn.raw)
        print(conn.src)
        print(conn.dst)
        print(conn.err)
        conn.push(conn.raw)
    end)
```