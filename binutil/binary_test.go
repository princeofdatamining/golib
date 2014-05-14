
package binutil_test

import (
    "testing"
    "github.com/princeofdatamining/golib/stream"
    "github.com/princeofdatamining/golib/binutil"
)

type byteorderRead struct {
    order   binutil.ByteOrder2
    b       byte
    i       uint32
}
type byteorderSample struct {
    order   binutil.ByteOrder2
    output  []byte
    reads   []*byteorderRead
}
var byteorderSamples = []*byteorderSample{
    {
        order: binutil.Fixed,
        output: []byte{ 0x09, 0x10, 0x13, 0x00, 0x00 },
        reads: []*byteorderRead{
            { binutil.Fixed  , 0x09, 0x00001310, },
            { binutil.FixedBE, 0x09, 0x10130000, },
        },
    },
    {
        order: binutil.FixedBE,
        output: []byte{ 0x09, 0x00, 0x00, 0x13, 0x10 },
        reads: []*byteorderRead{
            { binutil.Fixed  , 0x09, 0x10130000, },
            { binutil.FixedBE, 0x09, 0x00001310, },
        },
    },
}

func testByteOrder(t *testing.T, s *stream.Stream, in *byteorderSample) {
    s.SetSize(256)
    s.Seek(0, 0)
    data := s.Buf()
    whint := in.order.String()
    var m, n int
    m = in.order.PutUint8 (data, byte(0x09))
    n = in.order.PutUint32(data[m:], uint32(0x1310))
    n += m
    L := len(in.output)
    if n != L {
        t.Fatalf("%s: outsize must %v, but got %v\n", whint, L, n)
    }
    for i, b := range in.output {
        if b != data[i] {
            t.Fatalf("%s: output[%d] must %2X, but got %2X\n", whint, i, b, data[i])
        }
    }
    for _, read := range in.reads {
        s.Seek(0, 0)
        rhint := read.order.String()
        var b   byte
        var i   uint32
        b, m = read.order.Uint8(data)
        i, n = read.order.Uint32(data[m:])
        n += m
        if n != L {
            t.Errorf("%s: read by %s: final position must be %v, but got %v\n", whint, rhint, L, n)
        } else if b != read.b {
            t.Errorf("%s: read by %s: expect %.2X, but got %.2X\n", whint, rhint, read.b, b)
        } else if i != read.i {
            t.Errorf("%s: read by %s: expect %.8X, but got %.8X\n", whint, rhint, read.i, i)
        }
    }
}

func TestByteOrder(t *testing.T) {
    s := stream.NewStream(0)
    for _, in := range byteorderSamples {
        testByteOrder(t, s, in)
    }
}
