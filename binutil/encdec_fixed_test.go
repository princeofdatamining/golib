
package binutil_test

import (
    "reflect"
    "fmt"
    "testing"
    "github.com/princeofdatamining/golib/stream"
    "github.com/princeofdatamining/golib/binutil"
)

type point struct {
    X, Y    int
}
var (
    pA = point{   0,   0, }
    pB = point{ 800,   0, }
    pC = point{ 800, 600, }
    pD = point{   0, 600, }
)
type datatype struct {
    Yes     bool
    No      bool
    Int8    int8
    Int16   int16
    Int32   int32
    Int64   int64
    Byte    byte
    Word    uint16
    Long    uint32
    Quad    uint64
    Int     int
    Uint    uint
    Float   float32
    Double  float64
    Text    string
    Array   [4]point
    Points  []point
}
var data = &datatype{
    Yes:    true,
    Int8:   -1,
    Int16:  -26368,
    Int32:  -2004348655,
    Int64:  -6148895925951734307,
    Byte:   0xFF,               // 255
    Word:   0x9900,             // 39168
    Long:   0x88881111,         // 2290618641
    Quad:   0xAAAABBBBCCCCDDDD, // 12297848147757817309
    Int:    -65536,
    Uint:   0xFFFF0000,         // 4294901760
    Float:  1.0/9,
    Double: 1.0/9,
    Text:   "你好",
    Array:  [4]point{ pA, pB, pC, pD, },
    Points: []point { pA,     pC,     },
}
var (
    out_le_4 = []byte{
        1, 0,                                                                                           // 0 +2
        0xFF, 0x00, 0x99, 0x11, 0x11, 0x88, 0x88, 0xDD, 0xDD, 0xCC, 0xCC, 0xBB, 0xBB, 0xAA, 0xAA,       // 2 +15
        0xFF, 0x00, 0x99, 0x11, 0x11, 0x88, 0x88, 0xDD, 0xDD, 0xCC, 0xCC, 0xBB, 0xBB, 0xAA, 0xAA,       // 17+15
        0x00, 0x00, 0xFF, 0xFF, 0x00, 0x00, 0xFF, 0xFF,                                                 // 32+8
        0x39, 0x8E, 0xE3, 0x3D, 0x1C, 0xC7, 0x71, 0x1C, 0xC7, 0x71, 0xBC, 0x3F,                         // 40+12
        6, 0, 0, 0, 0xE4, 0xBD, 0xA0, 0xE5, 0xA5, 0xBD,                                                 // 52+4+6
        0x00, 0x00, 0, 0, 0x00, 0x00, 0, 0, 0x20, 0x03, 0, 0, 0x00, 0x00, 0, 0,                         // 62+16
        0x20, 0x03, 0, 0, 0x58, 0x02, 0, 0, 0x00, 0x00, 0, 0, 0x58, 0x02, 0, 0,                         // 78+16
        2, 0, 0, 0,                                                                                     // 94+4
        0x00, 0x00, 0, 0, 0x00, 0x00, 0, 0, 0x20, 0x03, 0, 0, 0x58, 0x02, 0, 0,                         // 98+16
    }
    out_be_4 = []byte{
        1, 0,                                                                                           // 0 +2
        0xFF, 0x99, 0x00, 0x88, 0x88, 0x11, 0x11, 0xAA, 0xAA, 0xBB, 0xBB, 0xCC, 0xCC, 0xDD, 0xDD,       // 2 +15
        0xFF, 0x99, 0x00, 0x88, 0x88, 0x11, 0x11, 0xAA, 0xAA, 0xBB, 0xBB, 0xCC, 0xCC, 0xDD, 0xDD,       // 17+15
        0xFF, 0xFF, 0x00, 0x00, 0xFF, 0xFF, 0x00, 0x00,                                                 // 32+8
        0x3D, 0xE3, 0x8E, 0x39, 0x3F, 0xBC, 0x71, 0xC7, 0x1C, 0x71, 0xC7, 0x1C,                         // 40+12
        0, 0, 0, 6, 0xE4, 0xBD, 0xA0, 0xE5, 0xA5, 0xBD,                                                 // 52+4+6
        0, 0, 0x00, 0x00, 0, 0, 0x00, 0x00, 0, 0, 0x03, 0x20, 0, 0, 0x00, 0x00,                         // 62+16
        0, 0, 0x03, 0x20, 0, 0, 0x02, 0x58, 0, 0, 0x00, 0x00, 0, 0, 0x02, 0x58,                         // 78+16
        0, 0, 0, 2,                                                                                     // 94+4
        0, 0, 0x00, 0x00, 0, 0, 0x00, 0x00, 0, 0, 0x03, 0x20, 0, 0, 0x02, 0x58,                         // 98+16
    }
    out_le_8 = []byte{
        1, 0,                                                                                           // 0 +2
        0xFF, 0x00, 0x99, 0x11, 0x11, 0x88, 0x88, 0xDD, 0xDD, 0xCC, 0xCC, 0xBB, 0xBB, 0xAA, 0xAA,       // 2 +15
        0xFF, 0x00, 0x99, 0x11, 0x11, 0x88, 0x88, 0xDD, 0xDD, 0xCC, 0xCC, 0xBB, 0xBB, 0xAA, 0xAA,       // 17+15
        0x00, 0x00, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x00, 0x00, 0xFF, 0xFF, 0x00, 0x00, 0x00, 0x00, // 32+16
        0x39, 0x8E, 0xE3, 0x3D, 0x1C, 0xC7, 0x71, 0x1C, 0xC7, 0x71, 0xBC, 0x3F,                         // 48+12
        6, 0, 0, 0, 0, 0, 0, 0, 0xE4, 0xBD, 0xA0, 0xE5, 0xA5, 0xBD,                                     // 60+8+6
        0x00, 0x00, 0, 0, 0, 0, 0, 0, 0x00, 0x00, 0, 0, 0, 0, 0, 0,                                     // 74+16
        0x20, 0x03, 0, 0, 0, 0, 0, 0, 0x00, 0x00, 0, 0, 0, 0, 0, 0,                                     // 90+16
        0x20, 0x03, 0, 0, 0, 0, 0, 0, 0x58, 0x02, 0, 0, 0, 0, 0, 0,                                     //106+16
        0x00, 0x00, 0, 0, 0, 0, 0, 0, 0x58, 0x02, 0, 0, 0, 0, 0, 0,                                     //122+16
        2, 0, 0, 0, 0, 0, 0, 0,                                                                         //138+8
        0x00, 0x00, 0, 0, 0, 0, 0, 0, 0x00, 0x00, 0, 0, 0, 0, 0, 0,                                     //146+16
        0x20, 0x03, 0, 0, 0, 0, 0, 0, 0x58, 0x02, 0, 0, 0, 0, 0, 0,                                     //162+16
    }
    out_be_8 = []byte{
        1, 0,                                                                                           // 0 +2
        0xFF, 0x99, 0x00, 0x88, 0x88, 0x11, 0x11, 0xAA, 0xAA, 0xBB, 0xBB, 0xCC, 0xCC, 0xDD, 0xDD,       // 2 +15
        0xFF, 0x99, 0x00, 0x88, 0x88, 0x11, 0x11, 0xAA, 0xAA, 0xBB, 0xBB, 0xCC, 0xCC, 0xDD, 0xDD,       // 17+15
        0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF, 0xFF, 0x00, 0x00, // 32+16
        0x3D, 0xE3, 0x8E, 0x39, 0x3F, 0xBC, 0x71, 0xC7, 0x1C, 0x71, 0xC7, 0x1C,                         // 48+12
        0, 0, 0, 0, 0, 0, 0, 6, 0xE4, 0xBD, 0xA0, 0xE5, 0xA5, 0xBD,                                     // 60+8+6
        0, 0, 0, 0, 0, 0, 0x00, 0x00, 0, 0, 0, 0, 0, 0, 0x00, 0x00,                                     // 74+16
        0, 0, 0, 0, 0, 0, 0x03, 0x20, 0, 0, 0, 0, 0, 0, 0x00, 0x00,                                     // 90+16
        0, 0, 0, 0, 0, 0, 0x03, 0x20, 0, 0, 0, 0, 0, 0, 0x02, 0x58,                                     //106+16
        0, 0, 0, 0, 0, 0, 0x00, 0x00, 0, 0, 0, 0, 0, 0, 0x02, 0x58,                                     //122+16
        0, 0, 0, 0, 0, 0, 0, 2,                                                                         //138+8
        0, 0, 0, 0, 0, 0, 0x00, 0x00, 0, 0, 0, 0, 0, 0, 0x00, 0x00,                                     //146+16
        0, 0, 0, 0, 0, 0, 0x03, 0x20, 0, 0, 0, 0, 0, 0, 0x02, 0x58,                                     //162+16
    }
)
type marshalTestData struct {
    data    *datatype
    order   binutil.ByteOrder2
    intNs   int
    outsize int
    out     []byte
}
var marshalTests = []*marshalTestData{
    { data, binutil.Fixed  , 4, 114, out_le_4, },
    { data, binutil.FixedBE, 4, 114, out_be_4, },
    { data, binutil.Fixed  , 8, 178, out_le_8, },
    { data, binutil.FixedBE, 8, 178, out_be_8, },
}

func testMarshal(t *testing.T, s *stream.Stream, in *marshalTestData) {
    var dummy = &datatype{}
    var n int
    var err error

    // println("*** encoding ***")
    s.SetSize(0)
    w := binutil.NewEncoder2(s, in.order, in.intNs)
    hint := w.String()
    n, err = w.Write(in.data)
    if err != nil {
        t.Errorf("encode(%s) error: %v\n", hint, err)
    }
    if n != in.outsize {
        t.Errorf("encode(%s) size must %v, but got %v\n", hint, in.outsize, n)
    }
    n = len(in.out)
    if n != in.outsize {
        t.Errorf("encode(%s) len(test) must %v, but got %v\n", hint, in.outsize, n)
    }
    b := s.Buf()
    n = len(b)
    if n != in.outsize {
        t.Errorf("encode(%s) len(dump) must %v, but got %v\n", hint, in.outsize, n)
    }
    for i, v := range in.out {
        if v != b[i] {
            t.Fatalf("marshal(%s) out[%v] must %2X, but got %2X\n", hint, i, v, b[i])
        }
    }
    //*
    // println("*** decoding ***")
    s.Seek(0, 0)
    r := binutil.NewDecoder2(s, in.order, in.intNs)
    n, err = r.Read(dummy)
    if err != nil {
        t.Errorf("decode(%s) error: %v\n", hint, err)
    }
    if n != in.outsize {
        t.Errorf("decode(%s) size must %v, but got %v\n", hint, in.outsize, n)
    }
    if !DeepEqual(in.data, dummy) {
        t.Fatalf("decode(%s) DeepEqual false\n", hint)
    }
    //*/
}
//*
func DeepEqual(a, b interface{}) (bool) {
    x := reflect.ValueOf(a)
    y := reflect.ValueOf(b)
    return deepequal("", x, y)
}
func deepequal(prop string, x, y reflect.Value) (ok bool) {
    x = reflect.Indirect(x)
    y = reflect.Indirect(y)
    m, n := x.Type(), y.Type()
    if m != n {
        fmt.Printf("%s: Type() %s != %s\n", prop, m.String(), n.String())
        return
    }
    switch x.Kind() {
    case reflect.Struct:
        m, n := x.NumField(), y.NumField()
        if m != n {
            fmt.Printf("%s: NumField() %v != %v\n", prop, m, n)
            return
        }
        for i := 0; i < n; i++ {
            if !deepequal(fmt.Sprintf("%sstruct[%d] ", prop, i), x.Field(i), y.Field(i)) {
                return
            }
        }
    case reflect.Array:
        m, n := x.Len(), y.Len()
        if m != n {
            fmt.Printf("%s: Len() %v != %v\n", prop, m, n)
            return
        }
        for i := 0; i < n; i++ {
            if !deepequal(fmt.Sprintf("%sarray[%d] ", prop, i), x.Index(i), y.Index(i)) {
                return
            }
        }
    case reflect.Slice:
        if m, n := x.IsNil(), y.IsNil(); m != n {
            fmt.Printf("%s: IsNil() %v != %v\n", prop, m, n)
            return
        }
        m, n := x.Len(), y.Len()
        if m != n {
            fmt.Printf("%s: Len() %v != %v\n", prop, m, n)
            return
        }
        for i := 0; i < n; i++ {
            if !deepequal(fmt.Sprintf("%sslice[%d] ", prop, i), x.Index(i), y.Index(i)) {
                return
            }
        }
    default:
        if x.Interface() != y.Interface() {
            fmt.Printf("%s: Interface() not equal\n", prop)
            return
        }
    }
    return true
}
//*/
func TestMarshal(t *testing.T) () {
    s := stream.NewStream(0)
    for _, in := range marshalTests {
        testMarshal(t, s, in)
    }
}
