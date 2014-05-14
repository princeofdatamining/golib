
package binutil

import (
    "io"
    "math"
    "reflect"
    "errors"
    "github.com/princeofdatamining/golib/stream"
)

type Encoder interface {
    Write(data interface{}) (n int, err error)
    Reset(w io.WriteSeeker) ()
    GetError() (error)
    GetByteOrder() (ByteOrder2)
    GetIntNBytes() (int)
    String() (string)
}
func NewEncoder2(w io.WriteSeeker, order ByteOrder2, intNbytes int) (Encoder) {
    var (
        s *stream.Stream
        ref bool
    )
    if w == nil {
        ref = true
    } else if s, ref = w.(*stream.Stream); !ref {
        s = stream.NewStream(0)
    }
    return &encoder{
        w: w,
        s: s,
        ref: ref,
        order: order,
        intNs: intNbytes,
    }
}
func NewEncoder (w io.WriteSeeker, order ByteOrder2               ) (Encoder) {
    return NewEncoder2(w, order, 4)
}

type encoder struct {
    w       io.WriteSeeker
    s       *stream.Stream
    ref     bool
    err     error
    order   ByteOrder2
    intNs   int
}

func (this *encoder) GetByteOrder() (ByteOrder2) { return this.order }
func (this *encoder) GetIntNBytes() (int) { return this.intNs }
func (this *encoder) String() (string) { return IntOrderString(this.order, this.intNs) }
func (this *encoder) GetError() (error) { return this.err }

func (this *encoder) Reset(w io.WriteSeeker) () {
    o := this.w
    this.w = w
    this.err = nil

    if !this.ref {
        this.s.SetSize(0)
        return
    }
    if o == w {
        return
    }

    s, ref := w.(*stream.Stream)
    if !ref {
        s = stream.NewStream(0)
    }
    this.s = s
    this.ref = ref
}

func (this *encoder) Write(data interface{}) (n int, err error) {
    n = this.s.Position()

    v := reflect.ValueOf(data)
    write(this, v)

    n = this.s.Position() - n
    return n, this.err
}

func write(e *encoder, v reflect.Value) () {
    if e.err != nil {
        return
    }
    v = reflect.Indirect(v)
    k := v.Kind()
    // println(k.String())
    switch k {

    case reflect.Bool:
        var x byte = 0
        if v.Bool() {
            x = 1
        }
        e.uint8(x)

    case reflect.Int  :
        e.int  (int  (v.Int()))
    case reflect.Int8 :
        e.int8 (int8 (v.Int()))
    case reflect.Int16:
        e.int16(int16(v.Int()))
    case reflect.Int32:
        e.int32(int32(v.Int()))
    case reflect.Int64:
        e.int64(      v.Int() )

    case reflect.Uint  :
        e.uint  (uint  (v.Uint()))
    case reflect.Uint8 :
        e.uint8 (uint8 (v.Uint()))
    case reflect.Uint16:
        e.uint16(uint16(v.Uint()))
    case reflect.Uint32:
        e.uint32(uint32(v.Uint()))
    case reflect.Uint64:
        e.uint64(       v.Uint() )

    case reflect.Float32:
        e.uint32(math.Float32bits(float32(v.Float())))
    case reflect.Float64:
        e.uint64(math.Float64bits(v.Float()))

    case reflect.Complex64:
        x := v.Complex()
        e.uint32(math.Float32bits(float32(real(x))))
        e.uint32(math.Float32bits(float32(imag(x))))
    case reflect.Complex128:
        x := v.Complex()
        e.uint64(math.Float64bits(real(x)))
        e.uint64(math.Float64bits(imag(x)))

    case reflect.Struct:
        n := v.NumField()
        for i := 0; i < n && e.err == nil; i++ {
            write(e, v.Field(i))
        }
    case reflect.Array:
        n := v.Len()
        for i := 0; i < n && e.err == nil; i++ {
            write(e, v.Index(i))
        }
    case reflect.String:
        x := []byte( v.String() )
        e.int( len(x) )
        if e.err == nil {
            e.bytes(x)
        }
    case reflect.Slice:
        n := v.Len()
        e.int( n )
        for i := 0; i < n && e.err == nil; i++ {
            write(e, v.Index(i))
        }
        //
    default:
        e.setError(errors.New("bintuil.encode: invalid type " + k.String()))
    }
}

func (this *encoder) setError(e error) () {
    if this.err == nil && e != nil {
        this.err = e
    }
}
func (this *encoder) next(n int) (b []byte, e error) {
    e = this.err
    if e != nil {
        return
    }
    b, e = this.s.Next(n)
    this.setError(e)
    return 
}
func (this *encoder) back(n int) () {
}
func (this *encoder) flush(b []byte) () {
    if !this.ref && this.err == nil {
        _, e := this.w.Write(b)
        this.setError(e)
    }
}
func (this *encoder) uint8(v uint8) {
    b, e := this.next( this.order.SizeUint8 (v) )
    if e == nil { this.order.PutUint8 (b, v) }
    this.flush(b)
}
func (this *encoder) uint16(v uint16) {
    b, e := this.next( this.order.SizeUint16(v) )
    if e == nil { this.order.PutUint16(b, v) }
    this.flush(b)
}
func (this *encoder) uint32(v uint32) {
    b, e := this.next( this.order.SizeUint32(v) )
    if e == nil { this.order.PutUint32(b, v) }
    this.flush(b)
}
func (this *encoder) uint64(v uint64) {
    b, e := this.next( this.order.SizeUint64(v) )
    if e == nil { this.order.PutUint64(b, v) }
    this.flush(b)
}
func (this *encoder) uint  (v uint) {
    b, e := this.next( SizeUint(this.order, this.intNs, v) )
    if e == nil { PutUint(this.order, this.intNs, b, v) }
    this.flush(b)
}
func (this *encoder)  int  (v int) {
    b, e := this.next( SizeInt (this.order, this.intNs, v) )
    if e == nil { PutInt (this.order, this.intNs, b, v) }
    this.flush(b)
}
func (this *encoder)  int8 (v int8 ) {
    b, e := this.next( this.order.SizeInt8 (v) )
    if e == nil { this.order.PutInt8 (b, v) }
    this.flush(b)
}
func (this *encoder)  int16(v int16) {
    b, e := this.next( this.order.SizeInt16(v) )
    if e == nil { this.order.PutInt16(b, v) }
    this.flush(b)
}
func (this *encoder)  int32(v int32) {
    b, e := this.next( this.order.SizeInt32(v) )
    if e == nil { this.order.PutInt32(b, v) }
    this.flush(b)
}
func (this *encoder)  int64(v int64) {
    b, e := this.next( this.order.SizeInt64(v) )
    if e == nil { this.order.PutInt64(b, v) }
    this.flush(b)
}
func (this *encoder) bytes(b []byte) () {
    if (this.err != nil) {
        return
    }
    _, e := this.s.Write(b)
    this.setError(e)
    this.flush(b)
}
