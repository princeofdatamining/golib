
package binutil

import (
    "io"
    "math"
    "errors"
)

type WireType int

const (
    WireVarint WireType = iota  // 0
    WireFixed64                 // 1
    WireBytes                   // 2
    WireStartGroup              // 3 deprecated
    WireEndGroup                // 4 deprecated
    WireFixed32                 // 5
)

var (
    ErrTagNotPositive  = errors.New("protobuf: tag must be positive")
    ErrInvalidWireType = errors.New("protobuf: invalid wiretype")
    errInvalidVarint   = errors.New("protobuf: read varint fail")
    errInvalidFixed64  = errors.New("protobuf: read fixed64 fail")
    errInvalidFixed32  = errors.New("protobuf: read fixed32 fail")
    errInvalidBytes    = errors.New("protobuf: read bytes fail")
)

type ProtobufEncoder interface {
    Reset(wr io.Writer) ()
    GetByteOrder() (ByteOrder2)
    // varint: (u)int
    PutUint64  (tag int, v  uint64) (int, error)
    PutUint32  (tag int, v  uint32) (int, error)
    PutInt64   (tag int, v   int64) (int, error)
    PutInt32   (tag int, v   int32) (int, error)
    // varint: zigzag
    PutSint64  (tag int, v   int64) (int, error)
    PutSint32  (tag int, v   int32) (int, error)
    // varint: int
    PutBool    (tag int, v   bool ) (int, error)
    PutEnum    (tag int, v   int  ) (int, error)
    // fixed64
    PutFixed64 (tag int, v  uint64) (int, error)
    PutSfixed64(tag int, v   int64) (int, error)
    PutFloat64 (tag int, v float64) (int, error)
    // fixed32
    PutFixed32 (tag int, v  uint32) (int, error)
    PutSfixed32(tag int, v   int32) (int, error)
    PutFloat32 (tag int, v float32) (int, error)
    // bytes
    PutBytes   (tag int, v  []byte) (int, error)
    PutString  (tag int, v  string) (int, error)
}
func newProtobufEncoder2(wr io.Writer, bo ByteOrder2) (ProtobufEncoder) {
    b := make([]byte, 32)
    return &pbEncoder{
        wr      : wr,
        bo      : bo,
        tagwire : b[:16],
        buf     : b[16:],
    }
}
func NewProtobufEncoder(wr io.Writer) (ProtobufEncoder) {
    return newProtobufEncoder2(wr, Fixed)
}

type ProtobufDecoder interface {
    Reset(rd io.ReadSeeker) ()
    GetByteOrder() (ByteOrder2)
    // tag & wire
    Tag() (n, tag int, wire WireType, err error)
    Skip(wire WireType) (n int, err error)
    // varint: (u)int
    Uint64  () (n int, v  uint64, err error)
    Uint32  () (n int, v  uint32, err error)
    Int64   () (n int, v   int64, err error)
    Int32   () (n int, v   int32, err error)
    // varint: zigzag
    Sint64  () (n int, v   int64, err error)
    Sint32  () (n int, v   int32, err error)
    // varint: int
    Bool    () (n int, v   bool , err error)
    Enum    () (n int, v   int  , err error)
    // fixed64
    Fixed64 () (n int, v  uint64, err error)
    Sfixed64() (n int, v   int64, err error)
    Float64 () (n int, v float64, err error)
    // fixed32
    Fixed32 () (n int, v  uint32, err error)
    Sfixed32() (n int, v   int32, err error)
    Float32 () (n int, v float32, err error)
    // bytes
    BodyLen () (n int, v   int  , err error)
    GetBody ([]byte) (int, error)
    Bytes   () (n int, v  []byte, err error)
    String  () (n int, v  string, err error)
}
func newProtobufDecoder2(rd io.ReadSeeker, bo ByteOrder2) (ProtobufDecoder) {
    b := make([]byte, 32)
    return &pbDecoder{
        rd      : rd,
        bo      : bo,
        buf     : b[16:],
    }
}
func NewProtobufDecoder(rd io.ReadSeeker) (ProtobufDecoder) {
    return newProtobufDecoder2(rd, Fixed)
}


type pbEncoder struct {
    wr      io.Writer
    bo      ByteOrder2
    tagwire []byte
    buf     []byte
}
func (this *pbEncoder) Reset(wr io.Writer) () { this.wr = wr }
func (this *pbEncoder) GetByteOrder() (ByteOrder2) { return this.bo }
func (this *pbEncoder) putTagWith(tag int, wire WireType, n int, attach []byte) (sz int, err error) {
    if tag <= 0 {
        return 0, ErrTagNotPositive
    }
    switch wire {
    case WireVarint, WireFixed64, WireFixed32, WireBytes:
    default:
        return 0, ErrInvalidWireType
    }
    //
    var (
        k, m    int
        bytes   bool
    )
    if L := len(attach); L > 0 {
        bytes = true
        n = L
    } else if n <= 0 {
        return 0, nil
    } else {
        attach = this.buf[:n]
    }
    //
    if u := (tag<<3) | int(wire); true {
        m = Varint.PutUint64(this.tagwire, uint64(u))
        k, err = this.wr.Write(this.tagwire[:m])
        sz += k
        if err != nil {
            return 
        }
    }
    //
    if bytes {
        m = Varint.PutUint64(this.tagwire, uint64(n))
        k, err = this.wr.Write(this.tagwire[:m])
        sz += k
        if err != nil {
            return 
        }
    }
    //
    if true {
        k, err = this.wr.Write(attach)
        sz += k
        if err != nil {
            return 
        }
    }

    return 
}

func (this *pbEncoder) PutUint64  (tag int, v  uint64) (int, error) { return this.putTagWith(tag, WireVarint , Varint .PutUint64(this.buf, v), nil) }
func (this *pbEncoder) PutFixed64 (tag int, v  uint64) (int, error) { return this.putTagWith(tag, WireFixed64, this.bo.PutUint64(this.buf, v), nil) }
func (this *pbEncoder) PutFixed32 (tag int, v  uint32) (int, error) { return this.putTagWith(tag, WireFixed32, this.bo.PutUint32(this.buf, v), nil) }
func (this *pbEncoder) PutBytes   (tag int, v  []byte) (int, error) { return this.putTagWith(tag, WireBytes  , 0                            , v  ) }

func (this *pbEncoder) PutInt64   (tag int, v   int64) (int, error) { return this.PutUint64 (tag, uint64(v)) }
func (this *pbEncoder) PutUint32  (tag int, v  uint32) (int, error) { return this.PutUint64 (tag, uint64(v)) }
func (this *pbEncoder) PutInt32   (tag int, v   int32) (int, error) { return this.PutUint32 (tag, uint32(v)) }

func (this *pbEncoder) PutSint64  (tag int, v   int64) (int, error) { return this.PutUint64 (tag, encodeZigzag(uint64(v))) }
func (this *pbEncoder) PutSint32  (tag int, v   int32) (int, error) { return this.PutUint64 (tag, encodeZigzag(uint64(v))) }

func (this *pbEncoder) putInt     (tag int, v   int  ) (int, error) { return this.PutInt32  (tag, int32(v)) }
func (this *pbEncoder) PutBool    (tag int, v   bool ) (int, error) { return this.putInt    (tag, bool2int(v)) }
func (this *pbEncoder) PutEnum    (tag int, v   int  ) (int, error) { return this.putInt    (tag, v) }

func (this *pbEncoder) PutSfixed64(tag int, v   int64) (int, error) { return this.PutFixed64(tag, uint64(v)) }
func (this *pbEncoder) PutFloat64 (tag int, v float64) (int, error) { return this.PutFixed64(tag, math.Float64bits(v)) }

func (this *pbEncoder) PutSfixed32(tag int, v   int32) (int, error) { return this.PutFixed32(tag, uint32(v)) }
func (this *pbEncoder) PutFloat32 (tag int, v float32) (int, error) { return this.PutFixed32(tag, math.Float32bits(v)) }

func (this *pbEncoder) PutString  (tag int, v  string) (int, error) { return this.PutBytes  (tag, []byte(v)) }



type pbDecoder struct {
    rd      io.ReadSeeker
    bo      ByteOrder2
    buf     []byte
}
func (this *pbDecoder) Reset(rd io.ReadSeeker) () { this.rd = rd }
func (this *pbDecoder) GetByteOrder() (ByteOrder2) { return this.bo }

func (this *pbDecoder) getVarint () (n int, u uint64, err error) {
    var m, i, k int
    var b byte
    for {
        if m, err = this.rd.Read(this.buf[n:n+4]); err != nil {
            return 0, 0, err
        }
        if m <= 0 {
            return 0, 0, errInvalidVarint
        }
        for i, b = range this.buf[n:n+m] {
            if b < 0x80 {
                k = i+1
                n += k
                if k != m {
                    this.rd.Seek(int64(k-m), 1)
                }
                u, _ = Varint.Uint64(this.buf[:n])
                return n, u, nil
            }
        }
        n += m
    }
}
func (this *pbDecoder) getFixed64() (n int, u uint64, err error) {
    need := 8
    if n, err = this.rd.Read(this.buf[:need]); err != nil {
        return 0, 0, err
    }
    if n != need {
        return 0, 0, errInvalidFixed64
    }
    u, _ = this.bo.Uint64(this.buf[:n])
    return n, u, nil
}
func (this *pbDecoder) getFixed32() (n int, u uint32, err error) {
    need := 4
    if n, err = this.rd.Read(this.buf[:need]); err != nil {
        return 0, 0, err
    }
    if n != need {
        return 0, 0, errInvalidFixed32
    }
    u, _ = this.bo.Uint32(this.buf[:n])
    return n, u, nil
}
func (this *pbDecoder) getBytes  () (n int, b []byte, err error) {
    n, L, err := this.getVarint()
    if err != nil {
        return 0, nil, err
    }
    if L <= 0 {
        return 0, nil, errInvalidBytes
    }
    l := int(L)
    if l > len(this.buf) {
        this.buf = make([]byte, l)
    }
    m, err := this.rd.Read(this.buf[:l])
    if m != l {
        return 0, nil, errInvalidBytes
    }
    n += m
    return n, this.buf[:l], nil 
}

func (this *pbDecoder) Tag() (n, tag int, wire WireType, err error) {
    n, u, err := this.getVarint()
    if err != nil {
        return 
    }
    tag = int(u >> 3)
    wire = WireType(u & 7)
    if tag <= 0 {
        err = ErrTagNotPositive
        return 
    }
    switch wire {
    case WireVarint, WireFixed64, WireFixed32, WireBytes:
    default:
        err = ErrInvalidWireType
        return 
    }
    err = nil
    return 
}
func (this *pbDecoder) Skip(wire WireType) (n int, err error) {
    switch wire {
    case WireVarint : n  , _, err = this.getVarint()
    case WireFixed64: n=8; _, err = this.rd.Seek(int64(n), 1)
    case WireFixed32: n=4; _, err = this.rd.Seek(int64(n), 1)
    case WireBytes:
        var u uint64
        if n, u, err = this.getVarint(); err == nil {
            n += int(u)
            _, err = this.rd.Seek(int64(u), 1)
        }
    default:
        err = ErrInvalidWireType
    }
    return 
}

func (this *pbDecoder) Uint64  () (n int, v  uint64, err error) { return this.getVarint () }
func (this *pbDecoder) Fixed64 () (n int, v  uint64, err error) { return this.getFixed64() }
func (this *pbDecoder) Fixed32 () (n int, v  uint32, err error) { return this.getFixed32() }
func (this *pbDecoder) Bytes   () (n int, v  []byte, err error) { return this.getBytes  () }

func (this *pbDecoder) Int64   () (n int, v   int64, err error) { n, u, err := this.Uint64 (); v =        int64             (u) ; return }
func (this *pbDecoder) Uint32  () (n int, v  uint32, err error) { n, u, err := this.Uint64 (); v =       uint32             (u) ; return }
func (this *pbDecoder) Int32   () (n int, v   int32, err error) { n, u, err := this.Uint32 (); v =        int32             (u) ; return }

func (this *pbDecoder) Sint64  () (n int, v   int64, err error) { n, u, err := this.Uint64 (); v =        int64(decodeZigzag(u)); return }
func (this *pbDecoder) Sint32  () (n int, v   int32, err error) { n, u, err := this.Sint64 (); v =        int32             (u) ; return }

func (this *pbDecoder) getInt  () (n int, v   int  , err error) { n, u, err := this.Int32  (); v =        int               (u) ; return }
func (this *pbDecoder) Bool    () (n int, v   bool , err error) { n, u, err := this.getInt (); v =        int2bool          (u) ; return }
func (this *pbDecoder) Enum    () (n int, v   int  , err error) { return this.getInt () }

func (this *pbDecoder) BodyLen () (n int, v   int  , err error) { return this.getInt () }
func (this *pbDecoder) GetBody (buf []byte) ( int  ,     error) { return this.rd.Read(buf) }

func (this *pbDecoder) Sfixed64() (n int, v   int64, err error) { n, u, err := this.Fixed64(); v =        int64             (u) ; return }
func (this *pbDecoder) Float64 () (n int, v float64, err error) { n, u, err := this.Fixed64(); v = math.Float64frombits     (u) ; return }

func (this *pbDecoder) Sfixed32() (n int, v   int32, err error) { n, u, err := this.Fixed32(); v =        int32             (u) ; return }
func (this *pbDecoder) Float32 () (n int, v float32, err error) { n, u, err := this.Fixed32(); v = math.Float32frombits     (u) ; return }

func (this *pbDecoder) String  () (n int, v  string, err error) { n, u, err := this.Bytes  (); v =        string            (u) ; return }
