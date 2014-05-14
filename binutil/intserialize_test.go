
package binutil_test

import (
    "fmt"
    "github.com/princeofdatamining/golib/binutil"
    "testing"
)

var bytes = make([]byte, 16)
func zerobytes(b []byte) {
    for i := range b {
        b[i] = 0
    }
}

type IntWRTestData struct {
    order   binutil.ByteOrder2
    nbytes  int
    Bytes   []byte
    u       uint
    i       int
}
var IntWRTests = []*IntWRTestData{
    { binutil.Fixed  , 4, []byte{0xff, 0xff, 0xff, 0xff}, 0xFFFFFFFF, -1, },
    { binutil.Fixed  , 4, []byte{0x01, 0x00, 0x00, 0x80}, 0x80000001, -2147483647, },
    { binutil.FixedBE, 4, []byte{0xff, 0xff, 0xff, 0xff}, 0xFFFFFFFF, -1, },
    { binutil.FixedBE, 4, []byte{0x80, 0x00, 0x00, 0x01}, 0x80000001, -2147483647, },
    /*
    { binutil.Fixed  , 8, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, 0xFFFFFFFFFFFFFFFF, -1, },
    { binutil.Fixed  , 8, []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x80}, 0x8000000000000001, -9223372036854775807, },
    { binutil.FixedBE, 8, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, 0xFFFFFFFFFFFFFFFF, -1, },
    { binutil.FixedBE, 8, []byte{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}, 0x8000000000000001, -9223372036854775807, },
    //*/
}
func testInt(t *testing.T, in *IntWRTestData) {
    s := fmt.Sprintf("%s:%d", in.order.String(), in.nbytes)
    var (
        i   int
        u   uint
    )
    //
    i, _ = binutil.GetInt(in.order, in.nbytes, in.Bytes)
    if i != in.i {
        t.Fatalf("[%s] Int() want %v, but got %v\n", s, in.i, i)
    }
    zerobytes(bytes)
    binutil.PutInt(in.order, in.nbytes, bytes, in.i)
    for i, b := range in.Bytes {
        if v := bytes[i]; b != v {
            t.Fatalf("[%s] PutInt(%v) want [%d]=%2X, but got %2X\n", s, in.i, i, b, v)
        }
    }
    //
    u, _ = binutil.GetUint(in.order, in.nbytes, in.Bytes)
    if u != in.u {
        t.Fatalf("[%s] Uint() want %X, but got %X\n", s, in.u, u)
    }
    zerobytes(bytes)
    binutil.PutUint(in.order, in.nbytes, bytes, in.u)
    for i, b := range in.Bytes {
        if v := bytes[i]; b != v {
            t.Fatalf("[%s] PutUint(%v) want [%d]=%2X, but got %2X\n", s, in.u, i, b, v)
        }
    }
}
func TestIntSerialize(t *testing.T) () {
    for _, in := range IntWRTests {
        testInt(t, in)
    }
}
