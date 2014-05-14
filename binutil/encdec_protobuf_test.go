
package binutil_test

import (
    "testing"
    "github.com/princeofdatamining/golib/stream"
    "github.com/princeofdatamining/golib/binutil"
)

var (
    c32s  int32 = -260
    c64s  int64 = -6
    c32u = uint32(c32s)
    c64u = uint64(c64s)
    c32f float32 = 1.0/9
    c64f float64 = 1.0/9
    cbool = true
    cstr = "123"
    cenum int = 30
    cx, cy int32 = 0x01ff, 0x08ff

    pb_size = 6+11+3+2+5+9+6+11+5+9+6+10+3+6+3+9+9
    pb_data = []byte{
        0x08, 0xFC, 0xFD, 0xFF, 0xFF, 0x0F,                              //  0+6  int32:varint
        0x10, 0xFA, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x01,//  6+11 int64:varint
        0x18, 0x87, 0x04,                                                // 17+3  sint32:zigzag32
        0x20, 0x0B,                                                      // 20+2  sint64:zigzag64
        0x2D, 0xFC, 0xFE, 0xFF, 0xFF,                                    // 22+5  sfixed32:fixed32
        0x31, 0xFA, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,            // 27+9  sfixed64:fixed64
        0x50, 0xFC, 0xFD, 0xFF, 0xFF, 0x0F,                              // 36+6  uint32:varint
        0x58, 0xFA, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x01,// 42+11 uint64:varint
        0x65, 0xFC, 0xFE, 0xFF, 0xFF,                                    // 53+5  fixed32:fixed32
        0x69, 0xFA, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,            // 58+9  fixed64:fixed64
        0xA5, 0x01, 0x39, 0x8E, 0xE3, 0x3D,                              // 67+6  float:fixed32
        0xA9, 0x01, 0x1C, 0xC7, 0x71, 0x1C, 0xC7, 0x71, 0xBC, 0x3F,      // 73+10 double:fixed64
        0xF0, 0x01, 0x01,                                                // 83+3  bool:varint
        0xFA, 0x01, 0x03, 0x31, 0x32, 0x33,                              // 86+6  string
        0x80, 0x02, 0x1E,                                                // 92+3  enum:varint
        0xC2, 0x02, 0x06, 0x08, 0xFF, 0x03, 0x10, 0xFF, 0x11,            // 95+9  message
        0xC2, 0x02, 0x06, 0x08, 0xFF, 0x11, 0x10, 0xFF, 0x03,            //104+9  message
    }
)
var (
    rd = stream.NewStreamFrom(pb_data)
    wr = stream.NewStream(0)
    dm = stream.NewStream(0)
    comp = new(pbDoor)
)
type pbPos struct {
    X, Y    int32
}
type pbDoor struct {
    I32, S32, D32   int32
    I64, S64, D64   int64
    U32, X32        uint32
    U64, X64        uint64
    F32             float32
    F64             float64
    B               bool
    S               string
    E               int
    Poslist         [2]pbPos
}
func init_pbDoor(Sour *pbDoor) {
  Sour.I32, Sour.S32, Sour.D32 = c32s, c32s, c32s
  Sour.I64, Sour.S64, Sour.D64 = c64s, c64s, c64s
  Sour.U32, Sour.X32 = c32u, c32u
  Sour.U64, Sour.X64 = c64u, c64u
  Sour.F32, Sour.F64 = c32f, c64f
  Sour.B = cbool
  Sour.S = cstr
  Sour.E = cenum
  Sour.Poslist[0].X, Sour.Poslist[0].Y = cx, cy
  Sour.Poslist[1].X, Sour.Poslist[1].Y = cy, cx
}
func comp_pbDoor(t *testing.T, door *pbDoor, s string) {
    switch {
    case comp.I32 != door.I32 : t.Fatalf(s, "I32 fail")
    case comp.S32 != door.S32 : t.Fatalf(s, "S32 fail")
    case comp.D32 != door.D32 : t.Fatalf(s, "D32 fail")
    case comp.I64 != door.I64 : t.Fatalf(s, "I64 fail")
    case comp.S64 != door.S64 : t.Fatalf(s, "S64 fail")
    case comp.D64 != door.D64 : t.Fatalf(s, "D64 fail")
    case comp.U32 != door.U32 : t.Fatalf(s, "U32 fail")
    case comp.X32 != door.X32 : t.Fatalf(s, "X32 fail")
    case comp.U64 != door.U64 : t.Fatalf(s, "U64 fail")
    case comp.X64 != door.X64 : t.Fatalf(s, "X64 fail")
    case comp.F32 != door.F32 : t.Fatalf(s, "F32 fail")
    case comp.F64 != door.F64 : t.Fatalf(s, "F64 fail")
    case comp.B   != door.B   : t.Fatalf(s, "B   fail")
    case comp.S   != door.S   : t.Fatalf(s, "S   fail")
    case comp.E   != door.E   : t.Fatalf(s, "E   fail")
    case comp.Poslist[0].X != door.Poslist[0].X : t.Fatalf(s, "Poslist[0].X fail")
    case comp.Poslist[0].Y != door.Poslist[0].Y : t.Fatalf(s, "Poslist[0].Y fail")
    case comp.Poslist[1].X != door.Poslist[1].X : t.Fatalf(s, "Poslist[1].X fail")
    case comp.Poslist[1].Y != door.Poslist[1].Y : t.Fatalf(s, "Poslist[1].Y fail")
    }
}
func comp_bin(t *testing.T, buff []byte, off, L int, f, s string) {
    for i := 0; i < L; i++ {
        if buff[off+i] != pb_data[off+i] {
            t.Fatalf(f+"  want % 2X\n  curr % 2X\n", s, pb_data[off:off+L], buff[off:off+L])
        }
    }
}
func decode_pos(t *testing.T, rd *stream.Stream, last int, pos *pbPos) {
    rd.Seek(0, 0)
    d := binutil.NewProtobufDecoder(rd)
    for ; rd.Position() < last; {
        _, tag, wire, _ := d.Tag()
        switch tag {
        case 0: t.Fatalf("test protobuf decode: got tag 0 when parse pos\n")
        case 1: _, pos.X, _ = d.Int32()
        case 2: _, pos.Y, _ = d.Int32()
        default: d.Skip(wire)
        }
    }
    if rd.Position() != last {
        t.Fatalf("test protobuf decode: parse pos, error tail position\n")
    }
}
func decode_door(t *testing.T, rd *stream.Stream, last int, door *pbDoor) {
    rd.Seek(0, 0)
    d := binutil.NewProtobufDecoder(rd)
    posidx := 0
    for ; rd.Position() < last; {
        _, tag, wire, _ := d.Tag()
        switch tag {
        case 0 :  t.Fatalf("test protobuf decode: got tag 0 when parse door\n")
        case 1 :  _, door.I32, _ = d.Int32()
        case 3 :  _, door.S32, _ = d.Sint32()
        case 5 :  _, door.D32, _ = d.Sfixed32()
        case 2 :  _, door.I64, _ = d.Int64()
        case 4 :  _, door.S64, _ = d.Sint64()
        case 6 :  _, door.D64, _ = d.Sfixed64()

        case 10:  _, door.U32, _ = d.Uint32()
        case 12:  _, door.X32, _ = d.Fixed32()
        case 11:  _, door.U64, _ = d.Uint64()
        case 13:  _, door.X64, _ = d.Fixed64()

        case 20:  _, door.F32, _ = d.Float32()
        case 21:  _, door.F64, _ = d.Float64()

        case 30:  _, door.B  , _ = d.Bool()
        case 31:  _, door.S  , _ = d.String()
        case 32:  _, door.E  , _ = d.Enum()
        case 40:
            _, u, _ := d.Uint64()
            m, L := int64(u), int(u)
            dm.SetSize(m)
            d.GetBody(dm.Buf()[:m])
            decode_pos(t, dm, L, &door.Poslist[posidx])
            if dm.Position() != L {
                t.Fatalf("test protobuf decode: got pos fail\n")
            }
            posidx++
        default: d.Skip(wire)
        }
    }
    if rd.Position() != last {
        t.Fatalf("test protobuf decode: parse pos, error tail position\n")
    }
}
func testPbDecode(t *testing.T) {
    door := new(pbDoor)
    decode_door(t, rd, pb_size, door)
    comp_pbDoor(t, door, "test protobuf decode: %s\n")
}
func encode_pos(wr *stream.Stream, pos *pbPos) {
    wr.SetSize(0)
    wr.Seek(0, 0)
    e := binutil.NewProtobufEncoder(wr)
    e.PutInt32(1, pos.X)
    e.PutInt32(2, pos.Y)
}
func encode_door(t *testing.T, wr *stream.Stream) {
    wr.SetSize(0)
    wr.Seek(0, 0)
    e := binutil.NewProtobufEncoder(wr)
    door := comp
    e.PutInt32(1, door.I32)
    e.PutInt64(2, door.I64)
    e.PutSint32(3, door.S32)
    e.PutSint64(4, door.S64)
    e.PutSfixed32(5, door.D32)
    e.PutSfixed64(6, door.D64)
    e.PutUint32(10, door.U32)
    e.PutUint64(11, door.U64)
    e.PutFixed32(12, door.X32)
    e.PutFixed64(13, door.X64)
    e.PutFloat32(20, door.F32)
    e.PutFloat64(21, door.F64)
    e.PutBool(30, door.B)
    e.PutString(31, door.S)
    e.PutEnum(32, door.E)
    encode_pos(dm, &door.Poslist[0]); e.PutBytes(40, dm.Buf()[:dm.Position()])
    encode_pos(dm, &door.Poslist[1]); e.PutBytes(40, dm.Buf()[:dm.Position()])
}
func testPbEncode(t *testing.T) {
    encode_door(t, wr)
    if wr.Position() != pb_size {
        t.Fatalf("test protobuf encode: size error\n")
    }
    buff := wr.Buf()
    f := "test protobuf encode: %s\n"
    comp_bin(t, buff,   0,6 , f, "I32 fail")
    comp_bin(t, buff,   6,11, f, "I64 fail")
    comp_bin(t, buff,  17,3 , f, "S32 fail")
    comp_bin(t, buff,  20,2 , f, "S64 fail")
    comp_bin(t, buff,  22,5 , f, "D32 fail")
    comp_bin(t, buff,  27,9 , f, "D64 fail")
    comp_bin(t, buff,  36,6 , f, "U32 fail")
    comp_bin(t, buff,  42,11, f, "U64 fail")
    comp_bin(t, buff,  53,5 , f, "X32 fail")
    comp_bin(t, buff,  58,9 , f, "X64 fail")
    comp_bin(t, buff,  67,6 , f, "F32 fail")
    comp_bin(t, buff,  73,10, f, "F64 fail")
    comp_bin(t, buff,  83,3 , f, "B   fail")
    comp_bin(t, buff,  86,6 , f, "S   fail")
    comp_bin(t, buff,  92,3 , f, "E   fail")
    comp_bin(t, buff,  95,9 , f, "poslist[0] fail")
    comp_bin(t, buff, 104,9 , f, "poslist[1] fail")
}
func TestProtobufEncDec(t *testing.T) {
    init_pbDoor(comp)
    testPbDecode(t)
    testPbEncode(t)
}
