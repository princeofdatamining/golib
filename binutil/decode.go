
package binutil

import (
    "io"
    "math"
    "reflect"
    "errors"
    "github.com/princeofdatamining/golib/stream"
)

type Decoder interface {
    Read (data interface{}) (n int, err error)
    Reset(r io. ReadSeeker) ()
    GetError() (error)
    GetByteOrder() (ByteOrder2)
    GetIntNBytes() (int)
    String() (string)
}
func NewDecoder2(r io. ReadSeeker, order ByteOrder2, intNbytes int) (Decoder) {
    var (
        s *stream.Stream
        ref bool
    )
    if r == nil {
        ref = true
    } else if s, ref = r.(*stream.Stream); !ref {
        s = stream.NewStream(0)
    }
    return &decoder{
        r: r,
        s: s,
        ref: ref,
        order: order,
        intNs: intNbytes,
    }
}
func NewDecoder (r io. ReadSeeker, order ByteOrder2               ) (Decoder) {
    return NewDecoder2(r, order, 4)
}

type decoder struct {
    r       io. ReadSeeker
    s       *stream.Stream
    ref     bool
    err     error
    order   ByteOrder2
    intNs   int
}

func (this *decoder) GetByteOrder() (ByteOrder2) { return this.order }
func (this *decoder) GetIntNBytes() (int) { return this.intNs }
func (this *decoder) String() (string) { return IntOrderString(this.order, this.intNs) }
func (this *decoder) GetError() (error) { return this.err }

func (this *decoder) Reset(r io. ReadSeeker) () {
    o := this.r
    this.r = r
    this.err = nil

    if !this.ref {
        this.s.SetSize(0)
        return
    }
    if o == r {
        return
    }

    s, ref := r.(*stream.Stream)
    if !ref {
        s = stream.NewStream(0)
    }
    this.s = s
    this.ref = ref
}

func (this *decoder)  Read(data interface{}) (n int, err error) {
    n = this.s.Position()

    v := reflect.ValueOf(data)
    read(this, v, false)

    n = this.s.Position() - n
    return n, this.err
}

func  read(d *decoder, v reflect.Value, noset bool) () {
    if d.err != nil {
        return
    }
    v = reflect.Indirect(v)
    k := v.Kind()
    // println(k.String())
    switch k {

    case reflect.Bool:
        x := true
        if d.uint8() == 0 {
            x = false
        }
        if !noset { v.SetBool(x) }

    case reflect.Int  :
        x := d.int  ()
        if !noset { v.SetInt(int64(x)) }
    case reflect.Int8 :
        x := d.int8 ()
        if !noset { v.SetInt(int64(x)) }
    case reflect.Int16:
        x := d.int16()
        if !noset { v.SetInt(int64(x)) }
    case reflect.Int32:
        x := d.int32()
        if !noset { v.SetInt(int64(x)) }
    case reflect.Int64:
        x := d.int64()
        if !noset { v.SetInt(      x ) }

    case reflect.Uint  :
        x := d.uint  ()
        if !noset { v.SetUint(uint64(x)) }
    case reflect.Uint8 :
        x := d.uint8 ()
        if !noset { v.SetUint(uint64(x)) }
    case reflect.Uint16:
        x := d.uint16()
        if !noset { v.SetUint(uint64(x)) }
    case reflect.Uint32:
        x := d.uint32()
        if !noset { v.SetUint(uint64(x)) }
    case reflect.Uint64:
        x := d.uint64()
        if !noset { v.SetUint(       x ) }

    case reflect.Float32:
        x := math.Float32frombits(d.uint32())
        if !noset { v.SetFloat(float64(x)) }
    case reflect.Float64:
        x := math.Float64frombits(d.uint64())
        if !noset { v.SetFloat(        x ) }

    case reflect.Complex64:
        r, i := math.Float32frombits(d.uint32()), math.Float32frombits(d.uint32())
        x := complex( float64(r), float64(i) )
        if !noset { v.SetComplex(x) }
    case reflect.Complex128:
        r, i := math.Float64frombits(d.uint64()), math.Float64frombits(d.uint64())
        x := complex(         r ,         i  )
        if !noset { v.SetComplex(x) }

    case reflect.Struct:
        n := v.NumField()
        for i := 0; i < n && d.err == nil; i++ {
            read(d, v.Field(i), noset || !v.CanSet())
        }
    case reflect.Array:
        n := v.Len()
        for i := 0; i < n && d.err == nil; i++ {
            read(d, v.Index(i), noset)
        }
    case reflect.String:
        n := d.int()
        x := string( d.bytes(n) )
        if d.err == nil && !noset {
            v.SetString(x)
        }
    case reflect.Slice:
        n := d.int()
        x := reflect.MakeSlice(v.Type(), n, n)
        for i := 0; i < n && d.err == nil; i++ {
            read(d, x.Index(i), noset)
        }
        if !noset { v.Set( x ) }
    default:
        d.setError(errors.New("bintuil.decode: invalid type " + k.String()))
    }
}

func (this *decoder) setError(e error) (error) {
    if this.err == nil && e != nil {
        this.err = e
    }
    return e
}
func (this *decoder) peek(n int) (b []byte, e error) {
    e = this.err
    if e != nil {
        return
    }
    var m int
    if !this.ref {
        b, e = this.s.Next(n)
        if e != nil {
            m, e = this.r.Read(b)
            if m != n {
                b = b[:m]
                this.s.Seek(int64(m-n), 1)
            }
        }
    } else {
        m = this.s.Len()
        if n > m {
            n = m
        }
        b, e = this.s.Next(n)
    }
    this.setError(e)
    return 
}
func (this *decoder) back(n int) () {
    var m int64 = int64(n)
    if m <= 0 {
        return 
    }
    if !this.ref {
        this.r.Seek(-m, 1)
    }
    this.s.Seek(-m, 1)
}
func (this *decoder) flush() () {
}
func (this *decoder) uint8 () (v uint8 ) {
    b, e := this.peek(16)
    if e != nil {
        return 
    }
    v, n := this.order.Uint8 (b)
    if n <= 0 {
        this.setError(ErrReadIncomplete)
        return 
    }
    this.back(len(b)-n)
    return 
}
func (this *decoder) uint16() (v uint16) {
    b, e := this.peek(16)
    if e != nil {
        return 
    }
    v, n := this.order.Uint16(b)
    if n <= 0 {
        this.setError(ErrReadIncomplete)
        return 
    }
    this.back(len(b)-n)
    return 
}
func (this *decoder) uint32() (v uint32) {
    b, e := this.peek(16)
    if e != nil {
        return 
    }
    v, n := this.order.Uint32(b)
    if n <= 0 {
        this.setError(ErrReadIncomplete)
        return 
    }
    this.back(len(b)-n)
    return 
}
func (this *decoder) uint64() (v uint64) {
    b, e := this.peek(16)
    if e != nil {
        return 
    }
    v, n := this.order.Uint64(b)
    if n <= 0 {
        this.setError(ErrReadIncomplete)
        return 
    }
    this.back(len(b)-n)
    return 
}
func (this *decoder) uint  () (v uint  ) {
    b, e := this.peek(16)
    if e != nil {
        return 
    }
    v, n := GetUint(this.order, this.intNs, b)
    if n <= 0 {
        this.setError(ErrReadIncomplete)
        return 
    }
    this.back(len(b)-n)
    return 
}
func (this *decoder)  int  () (v  int  ) {
    b, e := this.peek(16)
    if e != nil {
        return 
    }
    v, n := GetInt (this.order, this.intNs, b)
    if n <= 0 {
        this.setError(ErrReadIncomplete)
        return 
    }
    this.back(len(b)-n)
    return 
}
func (this *decoder)  int8 () (v int8 ) {
    b, e := this.peek(16)
    if e != nil {
        return 
    }
    v, n := this.order.Int8 (b)
    if n <= 0 {
        this.setError(ErrReadIncomplete)
        return 
    }
    this.back(len(b)-n)
    return 
}
func (this *decoder)  int16() (v int16) {
    b, e := this.peek(16)
    if e != nil {
        return 
    }
    v, n := this.order.Int16(b)
    if n <= 0 {
        this.setError(ErrReadIncomplete)
        return 
    }
    this.back(len(b)-n)
    return 
}
func (this *decoder)  int32() (v int32) {
    b, e := this.peek(16)
    if e != nil {
        return 
    }
    v, n := this.order.Int32(b)
    if n <= 0 {
        this.setError(ErrReadIncomplete)
        return 
    }
    this.back(len(b)-n)
    return 
}
func (this *decoder)  int64() (v int64) {
    b, e := this.peek(16)
    if e != nil {
        return 
    }
    v, n := this.order.Int64(b)
    if n <= 0 {
        this.setError(ErrReadIncomplete)
        return 
    }
    this.back(len(b)-n)
    return 
}
func (this *decoder) bytes(n int) (b []byte) {
    if this.err != nil {
        return
    }
    b, _ = this.peek(n)
    return 
}
