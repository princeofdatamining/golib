
package socksutil

/* http://www.ietf.org/rfc/rfc1928.txt

   The SOCKS request is formed as follows:

        +----+-----+-------+------+----------+----------+
        |VER | CMD |  RSV  | ATYP | DST.ADDR | DST.PORT |
        +----+-----+-------+------+----------+----------+
        | 1  |  1  | X'00' |  1   | Variable |    2     |
        +----+-----+-------+------+----------+----------+

     Where:

          o  VER    protocol version: X'05'
          o  CMD
             o  CONNECT X'01'
             o  BIND X'02'
             o  UDP ASSOCIATE X'03'
          o  RSV    RESERVED
          o  ATYP   address type of following address
             o  IP V4 address: X'01'
             o  DOMAINNAME: X'03'
             o  IP V6 address: X'04'
          o  DST.ADDR       desired destination address
          o  DST.PORT desired destination port in network octet
             order

//*/

import (
    "net"
    "errors"
    "fmt"
    "encoding/binary"
    "strconv"
)

type SocksVer int

const (
    svUndefined SocksVer = 0
    Ver4   = svUndefined + 4
    Ver4a  = Ver4
    Ver5   = svUndefined + 5
)

type SocksCmd int

const (
    scUndefined SocksCmd = 0
    CmdConnect  = scUndefined + 1
    CmdBin      = scUndefined + 2
    CmdUDP      = scUndefined + 3
)

type AddrType int

const (
    undefined AddrType  = 0
    IPv4        = undefined + 1
    Domain      = undefined + 3
    IPv6        = undefined + 4
)

var (
    errfInvalidAddress = "rfc1928: address error %s %v"
    errfInvalidPort = "rfc1928: invalid port %s %v"
)

func CalcRawLen(tAddr AddrType, hostLen int) (n int, err error) {
    n = 1+2 // plus with port(2)
    switch tAddr {
    case IPv4:
        n += net.IPv4len
    case IPv6:
        n += net.IPv6len
    case Domain:
        n += 1 + hostLen
    default:
        err = errSocksAddress
    }
    return
}

func GetRawLen(raw []byte) (n int, err error) {
    return CalcRawLen( AddrType(raw[0]), int(raw[1]) )
}

func GetRawLenAt(raw []byte, off int) (n int, err error) {
    return GetRawLen(raw[off:])
}


func DecodeHostPort(raw []byte) (host string, port int, err error) {
    port_pos := len(raw)-2
    switch AddrType(raw[0]) {
    case IPv4, IPv6:
        host = net.IP(raw[1:port_pos]).String()
    case Domain:
        host = string(raw[2:port_pos])
    default:
        return "", 0, errSocksAddress
    }
    port = int(binary.BigEndian.Uint16(raw[port_pos:]))
    return
}

func EncodeHostPort(host string, port int, tAddr AddrType) ([]byte, error) {
    l := len(host)
    n, err := CalcRawLen(tAddr, l)
    if err != nil {
        return nil, err
    }
    raw := make([]byte, n)
    raw[0] = byte(tAddr)
    switch tAddr {
    case Domain:
        raw[1] = byte(l)
        copy(raw[2:n-2], host)
    default:
        ip := net.ParseIP(host)
        copy(raw[1:n-2], ip)
    }
    binary.BigEndian.PutUint16(raw[n-2:], uint16(port))
    return raw, nil
}

func EncodeAddr(addr string, tAddr AddrType) (raw []byte, err error) {
    var (
        host, text string
        port int
    )
    host, text, err = net.SplitHostPort(addr)
    if err != nil {
        return nil, errors.New(fmt.Sprintf(errfInvalidAddress, addr, err))
    }
    port, err = strconv.Atoi(text)
    if err != nil {
        return nil, errors.New(fmt.Sprintf(errfInvalidPort, text, err))
    }
    raw, err = EncodeHostPort(host, port, tAddr)
    return 
}
