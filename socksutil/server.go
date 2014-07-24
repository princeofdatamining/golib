
package socksutil

import (
    "fmt"
    "errors"
    "net"
    "io"
    "bytes"
    "encoding/binary"
    "strconv"
)

var (
    _ = fmt.Println

    errSocksVer     = errors.New("socks version not supported")
    errExtraData    = errors.New("socks handshake get extra data")
    errExtraReqData = errors.New("socks request get extra data")
    errSocksCommand = errors.New("socks command not supported")
    errSocksAddress = errors.New("socks address type not supported")
    errTooLongUser  = errors.New("socks userid too long")
    errTooLongDomain= errors.New("socks domain too long")
    errAuthenticate = errors.New("socks authenticate fail")
    errSocksAnswer  = errors.New("socks error while responsing")
)

//*

type AuthenticateFunc func(userid, password string) (bool, error)

func alwaysok(userid, password string) (bool, error) {
    return true, nil
}

var authFunc AuthenticateFunc = alwaysok

func SetAuthenticateFunc(f AuthenticateFunc) {
    authFunc = f
}

//*/

func ParseSocksRequest(conn net.Conn) (rawaddr []byte, host string, err error) {
    buf := make([]byte, 258)
    var n int
    // make sure we get 2 bytes
    if n, err = io.ReadAtLeast(conn, buf, 2); err != nil {
        return
    }
    // debug.Printf("first n shakehand bytes: %v\n", buf[:n])
    switch SocksVer(buf[0]) {
    case Ver4:
        return handleRequest4(conn, buf, n)
    case Ver5:
        nmethods := int(buf[1])
        msgLen := nmethods + 2
        if n > msgLen {
            err = errExtraData
            return
        } else if n < msgLen {
            if _, err = io.ReadFull(conn, buf[n:msgLen]); err != nil {
                return
            }
            // debug.Printf("handshake data: %v\n", buf[:msgLen])
        }
        // send confirmation: version 5, and no authentication required
        _, err = conn.Write( []byte{byte(Ver5), 0} )
        if err != nil {
            return
        }
        // handle request
        return handleRequest5(conn)
    default:
        err = errSocksVer
    }
    return
}

func RequestSocks5(conn net.Conn, host string, port int) (err error) {
    buf := make([]byte, 258)
    var n int
    // handshake
    buf[0] = byte(Ver5)
    buf[1] = 0 // nmethods
    if _, err = conn.Write(buf[:2]); err != nil {
        return 
    }
    if n, err = io.ReadAtLeast(conn, buf, 2); err != nil {
        return 
    }
    if n != 2 {
        err = errExtraData
        return 
    } else if buf[0] != byte(Ver5) || buf[1] != 0 {
        err = errSocksVer
        return 
    }
    // request
    buf[s5Version] = byte(Ver5)
    buf[s5Command] = byte(CmdConnect)
    buf[s5Reversed] = 0
    raw, err := EncodeHostPort(host, port, Domain)
    copy(buf[s5AddrType:], raw)
    if _, err = conn.Write(buf[:3+len(raw)]); err != nil {
        return 
    }
    if _, _, err = ParseSocks5Request(conn); err != nil {
        return 
    }
    return nil
}


const (
    s4Version   = 0
    s4Command   = 1
    s4Port      = 2
    s4IP        = 4
    s4User      = 8
    s4MaxStrLen = 124
)

func handleRequest4(conn net.Conn, buf []byte, n int) (rawaddr []byte, host string, err error) {
    // have read at least 2 bytes
    if buf[s4Version] != byte(Ver4) {
        err = errSocksVer
        return
    }
    if buf[s4Command] != byte(CmdConnect) {
        err = errSocksCommand
        return
    }
    var (
        read int
        port int
        addrType AddrType
        userid string
        userLen int
        domainPos int = s4User
        domainLen int
        tail int
    )

    // force read port(uint16) and address(ipv4)
    if n < s4User {
        if read, err = io.ReadAtLeast(conn, buf[n:], s4User-n); err != nil {
            return
        }
        n += read
    }
    port = int(binary.BigEndian.Uint16(buf[s4Port:s4IP]))
    host = net.IP(buf[s4IP:s4User]).String()
    // debug.Println(fmt.Sprintf("host: `%v`, port: `%v`", host, port))

    // get userid part
    if true {
        for {
            userLen = bytes.IndexByte(buf[s4User:n], 0)
            if userLen >= 0 {
                userid = string(buf[s4User:s4User+userLen])
                domainPos += userLen + 1
                tail = domainPos
                break
            }
            if read, err = io.ReadAtLeast(conn, buf[n:], 1); err != nil {
                return
            }
            n += read
            if n >= s4User+s4MaxStrLen {
                err = errTooLongUser
                return
            }
        }
    }
    if ok, _ := authFunc(userid, ""); !ok {
        err = errAuthenticate
        return
    }
    // debug.Println(fmt.Sprintf("userid: `%v`", userid))

    // possible domain part
    if host == "0.0.0.1" {
        addrType = Domain
        for {
            domainLen = bytes.IndexByte(buf[domainPos:n], 0)
            if domainLen >= 0 {
                host = string(buf[domainPos:domainPos+domainLen])
                tail += domainLen + 1
                break
            }
            if read, err = io.ReadAtLeast(conn, buf[n:], 1); err != nil {
                return
            }
            n += read
            if n >= s4User+s4MaxStrLen*2 {
                err = errTooLongDomain
                return
            }
        }
    }

    if n > tail {
        err = errExtraReqData
        return
    }

    rawaddr, err = EncodeHostPort(host, port, addrType)
    host = net.JoinHostPort(host, strconv.Itoa(port))
    // debug.Println(fmt.Sprintf("host: `%v`, rawaddr: `%v`", host, rawaddr))

    // Sending connection established message immediately to client.
    //                             OK    port  address(v4) 
    _, err = conn.Write([]byte{ 0, 0x5A, 0, 0, 0, 0, 0, 0 })
    if err != nil {
        // debug.Printf("send connection confirmation: %v\n", err)
    }

    return
}


const (
    s5Version   = 0
    s5Command   = 1
    s5Reversed  = 2
    s5AddrType  = 3
    s5AddrBody  = 4
)

func foo() () {
    //
}

func ParseSocks5Request(conn net.Conn) (rawaddr []byte, cmd byte, err error) {
    buf := make([]byte, 263)
    var n int
    // make sure we get possible domain length field
    if n, err = io.ReadAtLeast(conn, buf, s5AddrBody+1); err != nil {
        return
    }
    // debug.Printf("first n request bytes: %v\n", buf[:n])
    if buf[s5Version] != byte(Ver5) {
        err = errSocksVer
        return 
    }
    cmd = buf[s5Command]

    reqLen, err := GetRawLenAt(buf, s5AddrType)
    if err != nil {
        return
    }
    reqLen += s5AddrType

    if n > reqLen {
        err = errExtraReqData
        return
    } else if n < reqLen {
        if _, err = io.ReadFull(conn, buf[n:reqLen]); err != nil {
            return
        }
        // debug.Printf("request data: %v\n", buf[:reqLen])
    }

    return buf[s5AddrType:reqLen], cmd, nil
}

func handleRequest5(conn net.Conn) (rawaddr []byte, host string, err error) {
    var cmd byte
    if rawaddr, cmd, err = ParseSocks5Request(conn); err != nil {
        return 
    }
    if cmd != byte(CmdConnect) {
        err = errSocksCommand
        return 
    }
    host, port, err := DecodeHostPort(rawaddr)
    if err != nil {
        return
    }
    host = net.JoinHostPort(host, strconv.Itoa(port))

    // Sending connection established message immediately to client.
    //                                    OK  Reversed          address(v4) port
    _, err = conn.Write([]byte{ byte(Ver5), 0, 0, byte(IPv4), 0, 0, 0, 0, 0, 0 })
    if err != nil {
        // debug.Printf("send connection confirmation: %v\n", err)
    }

    return
}
