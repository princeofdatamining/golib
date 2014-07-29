
package netutil_test

import (
    "testing"
    "net"
    "reflect"
    "encoding/binary"
)

type testAtonIO struct {
    s   string
    b   []byte
    v   uint32
}
var testAtonIOs = []*testAtonIO{
    &testAtonIO{ "127.0.0.1", []byte{0x7f, 0x00, 0x00, 0x01                        }, 0x0100007f, },
    &testAtonIO{ "::1"      , []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},          0, },
}
func TestAton(t *testing.T) () {
    for _, io := range testAtonIOs {
        ip := net.ParseIP(io.s)
        if len(io.b) == net.IPv6len {
            ip = ip.To16()
        } else {
            ip = ip.To4()
        }
        if ip == nil {
            t.Fatalf("atob(%q) failed\n", io.s)
        }
        if !reflect.DeepEqual(io.b, []byte(ip)) {
            t.Fatalf("atob(%q) must `% 02X`, but got `% 02X`\n", io.s, io.b, []byte(ip))
        }
        if len(ip) != net.IPv6len {
            v := binary.LittleEndian.Uint32(ip)
            if io.v != v {
                t.Fatalf("aton(%q) must %08X, but got %08X\n", io.s, io.v, v)
            }
        }
    }
}
