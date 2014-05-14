
package binutil

import (
    "encoding/binary"
    "errors"
    "fmt"
)

type ByteOrder2 interface {
    SizeUint64(uint64) (int)
    SizeUint32(uint32) (int)
    SizeUint16(uint16) (int)
    SizeUint8 (uint8 ) (int)
    SizeInt64 ( int64) (int)
    SizeInt32 ( int32) (int)
    SizeInt16 ( int16) (int)
    SizeInt8  ( int8 ) (int)
    Uint64([]byte) (uint64, int)
    Uint32([]byte) (uint32, int)
    Uint16([]byte) (uint16, int)
    Uint8 ([]byte) (uint8 , int)
    Int64 ([]byte) ( int64, int)
    Int32 ([]byte) ( int32, int)
    Int16 ([]byte) ( int16, int)
    Int8  ([]byte) ( int8 , int)
    PutUint64([]byte, uint64) (int)
    PutUint32([]byte, uint32) (int)
    PutUint16([]byte, uint16) (int)
    PutUint8 ([]byte, uint8 ) (int)
    PutInt64 ([]byte,  int64) (int)
    PutInt32 ([]byte,  int32) (int)
    PutInt16 ([]byte,  int16) (int)
    PutInt8  ([]byte,  int8 ) (int)
    String() (string)
}

var (
    Varint  ByteOrder2 = varint{}
    Zigzag  ByteOrder2 = zigzag{}
    Fixed   ByteOrder2 = fixedle{}
    FixedBE ByteOrder2 = fixedbe{}

    ErrReadIncomplete = errors.New("binutil: read fail: incomplete")
)

//

func int64bits(intNbytes int) (bool) { return intNbytes >= 8 }
func IntOrderString(order ByteOrder2, intNbytes int) (string) {
    n := 4
    if int64bits(intNbytes) { n = 8 }
    return fmt.Sprintf("%s(u|i=%d)", order.String(), n)
}

func SizeUint(order ByteOrder2, intNbytes int,              v uint) (n int) {
    if int64bits(intNbytes) {
        return order.SizeUint64(uint64(v))
    } else {
        return order.SizeUint32(uint32(v))
    }
}
func SizeInt (order ByteOrder2, intNbytes int,              v  int) (n int) {
    if int64bits(intNbytes) {
        return order.SizeInt64 ( int64(v))
    } else {
        return order.SizeInt32 ( int32(v))
    }
}
func PutUint (order ByteOrder2, intNbytes int, buf []byte,  v uint) (n int) {
    if int64bits(intNbytes) {
        return order.PutUint64(buf, uint64(v))
    } else {
        return order.PutUint32(buf, uint32(v))
    }
}
func PutInt  (order ByteOrder2, intNbytes int, buf []byte,  v  int) (n int) {
    if int64bits(intNbytes) {
        return order.PutInt64 (buf,  int64(v))
    } else {
        return order.PutInt32 (buf,  int32(v))
    }
}
func GetUint (order ByteOrder2, intNbytes int, buf []byte) (v uint,  n int) {
    if int64bits(intNbytes) {
        u, n := order.Uint64(buf); return uint(u), n
    } else {
        u, n := order.Uint32(buf); return uint(u), n
    }
}
func GetInt  (order ByteOrder2, intNbytes int, buf []byte) (v  int,  n int) {
    if int64bits(intNbytes) {
        u, n := order.Int64 (buf); return  int(u), n
    } else {
        u, n := order.Int32 (buf); return  int(u), n
    }
}

//

var (
    fixedLE binary.ByteOrder = binary.LittleEndian
    fixedBE binary.ByteOrder = binary.BigEndian
)
func sizeVarint  (x uint64) (n int) {
    for ; x > 0x7F; x >>= 7 {
        n++
    }
    return n+1
}
func encodeZigzag(x uint64) (uint64) {
    return (x << 1) ^ uint64(int64(x)>>63)
}
func decodeZigzag(x uint64) (uint64) {
    msb := int64(x & 1)
    return (x >> 1) ^ uint64((msb<<63)>>63)
}
func bool2int(v bool) (int) { if v { return 1 }; return 0 }
func int2bool(v int) (bool) { if v == 0 { return false }; return true }

// LittleEndian

type fixedle struct{}
func (this fixedle) String() (string) { return "Fixed" }
func (this fixedle) SizeUint64(v uint64) (int) { return 8 }
func (this fixedle) SizeUint32(v uint32) (int) { return 4 }
func (this fixedle) SizeUint16(v uint16) (int) { return 2 }
func (this fixedle) SizeUint8 (v uint8 ) (int) { return 1 }
func (this fixedle) SizeInt64 (v  int64) (int) { return 8 }
func (this fixedle) SizeInt32 (v  int32) (int) { return 4 }
func (this fixedle) SizeInt16 (v  int16) (int) { return 2 }
func (this fixedle) SizeInt8  (v  int8 ) (int) { return 1 }
func (this fixedle) Uint64(buf []byte) (x uint64, n int) { 
    if n = this.SizeUint64(x); n > len(buf) {
        return 0, 0
    }
    return fixedLE.Uint64(buf), n
}
func (this fixedle) Uint32(buf []byte) (x uint32, n int) { 
    if n = this.SizeUint32(x); n > len(buf) {
        return 0, 0
    }
    return fixedLE.Uint32(buf), n
}
func (this fixedle) Uint16(buf []byte) (x uint16, n int) { 
    if n = this.SizeUint16(x); n > len(buf) {
        return 0, 0
    }
    return fixedLE.Uint16(buf), n
}
func (this fixedle) Uint8 (buf []byte) (x uint8 , n int) { 
    if n = this.SizeUint8 (x); n > len(buf) {
        return 0, 0
    }
    return buf[0], n
}
func (this fixedle)  Int64(buf []byte) (x  int64, n int) { u, n := this.Uint64(buf); return int64(u), n }
func (this fixedle)  Int32(buf []byte) (x  int32, n int) { u, n := this.Uint32(buf); return int32(u), n }
func (this fixedle)  Int16(buf []byte) (x  int16, n int) { u, n := this.Uint16(buf); return int16(u), n }
func (this fixedle)  Int8 (buf []byte) (x  int8 , n int) { u, n := this.Uint8 (buf); return int8 (u), n }
func (this fixedle) PutUint64(buf []byte, x uint64) (n int) {
    if n = this.SizeUint64(x); n > len(buf) {
        return 0
    }
    fixedLE.PutUint64(buf, x)
    return 
}
func (this fixedle) PutUint32(buf []byte, x uint32) (n int) {
    if n = this.SizeUint32(x); n > len(buf) {
        return 0
    }
    fixedLE.PutUint32(buf, x)
    return 
}
func (this fixedle) PutUint16(buf []byte, x uint16) (n int) {
    if n = this.SizeUint16(x); n > len(buf) {
        return 0
    }
    fixedLE.PutUint16(buf, x)
    return 
}
func (this fixedle) PutUint8 (buf []byte, x uint8 ) (n int) {
    if n = this.SizeUint8 (x); n > len(buf) {
        return 0
    }
    buf[0] = x
    return 
}
func (this fixedle)  PutInt64(buf []byte, x  int64) (n int) { return this.PutUint64(buf, uint64(x)) }
func (this fixedle)  PutInt32(buf []byte, x  int32) (n int) { return this.PutUint32(buf, uint32(x)) }
func (this fixedle)  PutInt16(buf []byte, x  int16) (n int) { return this.PutUint16(buf, uint16(x)) }
func (this fixedle)   PutInt8(buf []byte, x  int8 ) (n int) { return this.PutUint8 (buf, uint8 (x)) }

// BigEndian

type fixedbe struct{}
func (this fixedbe) String() (string) { return "FixedBE" }
func (this fixedbe) SizeUint64(v uint64) (int) { return 8 }
func (this fixedbe) SizeUint32(v uint32) (int) { return 4 }
func (this fixedbe) SizeUint16(v uint16) (int) { return 2 }
func (this fixedbe) SizeUint8 (v uint8 ) (int) { return 1 }
func (this fixedbe) SizeInt64 (v  int64) (int) { return 8 }
func (this fixedbe) SizeInt32 (v  int32) (int) { return 4 }
func (this fixedbe) SizeInt16 (v  int16) (int) { return 2 }
func (this fixedbe) SizeInt8  (v  int8 ) (int) { return 1 }
func (this fixedbe) Uint64(buf []byte) (x uint64, n int) { 
    if n = this.SizeUint64(x); n > len(buf) {
        return 0, 0
    }
    return fixedBE.Uint64(buf), n
}
func (this fixedbe) Uint32(buf []byte) (x uint32, n int) { 
    if n = this.SizeUint32(x); n > len(buf) {
        return 0, 0
    }
    return fixedBE.Uint32(buf), n
}
func (this fixedbe) Uint16(buf []byte) (x uint16, n int) { 
    if n = this.SizeUint16(x); n > len(buf) {
        return 0, 0
    }
    return fixedBE.Uint16(buf), n
}
func (this fixedbe) Uint8 (buf []byte) (x uint8 , n int) { 
    if n = this.SizeUint8 (x); n > len(buf) {
        return 0, 0
    }
    return buf[0], n
}
func (this fixedbe)  Int64(buf []byte) (x  int64, n int) { u, n := this.Uint64(buf); return int64(u), n }
func (this fixedbe)  Int32(buf []byte) (x  int32, n int) { u, n := this.Uint32(buf); return int32(u), n }
func (this fixedbe)  Int16(buf []byte) (x  int16, n int) { u, n := this.Uint16(buf); return int16(u), n }
func (this fixedbe)  Int8 (buf []byte) (x  int8 , n int) { u, n := this.Uint8 (buf); return int8 (u), n }
func (this fixedbe) PutUint64(buf []byte, x uint64) (n int) {
    if n = this.SizeUint64(x); n > len(buf) {
        return 0
    }
    fixedBE.PutUint64(buf, x)
    return 
}
func (this fixedbe) PutUint32(buf []byte, x uint32) (n int) {
    if n = this.SizeUint32(x); n > len(buf) {
        return 0
    }
    fixedBE.PutUint32(buf, x)
    return 
}
func (this fixedbe) PutUint16(buf []byte, x uint16) (n int) {
    if n = this.SizeUint16(x); n > len(buf) {
        return 0
    }
    fixedBE.PutUint16(buf, x)
    return 
}
func (this fixedbe) PutUint8 (buf []byte, x uint8 ) (n int) {
    if n = this.SizeUint8 (x); n > len(buf) {
        return 0
    }
    buf[0] = x
    return 
}
func (this fixedbe)  PutInt64(buf []byte, x  int64) (n int) { return this.PutUint64(buf, uint64(x)) }
func (this fixedbe)  PutInt32(buf []byte, x  int32) (n int) { return this.PutUint32(buf, uint32(x)) }
func (this fixedbe)  PutInt16(buf []byte, x  int16) (n int) { return this.PutUint16(buf, uint16(x)) }
func (this fixedbe)   PutInt8(buf []byte, x  int8 ) (n int) { return this.PutUint8 (buf, uint8 (x)) }

// Varint

type varint struct{}
func (this varint) String() (string) { return "Varint" }
func (this varint) SizeUint64(v uint64) (int) { return sizeVarint(       v ) }
func (this varint) SizeUint32(v uint32) (int) { return this.SizeUint64(uint64(v)) }
func (this varint) SizeUint16(v uint16) (int) { return this.SizeUint64(uint64(v)) }
func (this varint) SizeUint8 (v uint8 ) (int) { return this.SizeUint64(uint64(v)) }
func (this varint) SizeInt64 (v  int64) (int) { return this.SizeUint64(uint64(v)) }
func (this varint) SizeInt32 (v  int32) (int) { return this.SizeUint32(uint32(v)) }
func (this varint) SizeInt16 (v  int16) (int) { return this.SizeUint16(uint16(v)) }
func (this varint) SizeInt8  (v  int8 ) (int) { return this.SizeUint8 (uint8 (v)) }
func (this varint) Uint64(buf []byte) (x uint64, n int) {
    L := len(buf)
    for shift := uint(0); shift < 64; shift += 7 {
        if n >= L {
            return 0, 0
        }
        b := uint64(buf[n])
        n++
        x |= (b & 0x7F) << shift
        if (b & 0x80) == 0 {
            return x, n
        }
    }
    // The number is too large to represent in a 64-bit value.
    return 0, 0
}
func (this varint) Uint32(buf []byte) (x uint32, n int) {
    u, n := this.Uint64(buf)
    if n <= 0 || u > 0xFFFFFFFF {
        return 0, 0
    }
    return uint32(u), n
}
func (this varint) Uint16(buf []byte) (x uint16, n int) {
    u, n := this.Uint64(buf)
    if n <= 0 || u > 0xFFFF     {
        return 0, 0
    }
    return uint16(u), n
}
func (this varint)  Uint8(buf []byte) (x uint8 , n int) {
    u, n := this.Uint64(buf)
    if n <= 0 || u > 0xFF       {
        return 0, 0
    }
    return uint8(u), n
}
func (this varint)  Int64(buf []byte) (x  int64, n int) {
    u, n := this.Uint64(buf)
    return int64(u), n
}
func (this varint)  Int32(buf []byte) (x  int32, n int) {
    i, n := this.Int64(buf)
    if n <= 0 || i > 0x7FFFFFFF || i <= -0x80000000 {
        return 0, 0
    }
    return int32(i), n
}
func (this varint)  Int16(buf []byte) (x  int16, n int) {
    i, n := this.Int64(buf)
    if n <= 0 || i > 0x7FFF     || i <= -0x8000     {
        return 0, 0
    }
    return int16(i), n
}
func (this varint)   Int8(buf []byte) (x  int8 , n int) {
    i, n := this.Int64(buf)
    if n <= 0 || i > 0x7F       || i <= -0x80       {
        return 0, 0
    }
    return int8 (i), n
}
func (this varint) PutUint64(buf []byte, x uint64) (n int) {
    if this.SizeUint64(x) > len(buf) {
        return 0
    }
    for ; x > 0x7F; x >>= 7 {
        buf[n] = 0x80 | uint8(x & 0x7F)
        n++
    }
    buf[n] = uint8(x)
    return n+1
}
func (this varint) PutUint32(buf []byte, x uint32) (n int) { return this.PutUint64(buf, uint64(x)) }
func (this varint) PutUint16(buf []byte, x uint16) (n int) { return this.PutUint64(buf, uint64(x)) }
func (this varint)  PutUint8(buf []byte, x uint8 ) (n int) { return this.PutUint64(buf, uint64(x)) }
func (this varint)  PutInt64(buf []byte, x  int64) (n int) { return this.PutUint64(buf, uint64(x)) }
func (this varint)  PutInt32(buf []byte, x  int32) (n int) { return this.PutUint32(buf, uint32(x)) }
func (this varint)  PutInt16(buf []byte, x  int16) (n int) { return this.PutUint16(buf, uint16(x)) }
func (this varint)   PutInt8(buf []byte, x  int8 ) (n int) { return this.PutUint8 (buf, uint8 (x)) }

// Zigzag

type zigzag struct{}
func (this zigzag) String() (string) { return "Zigzag" }
func (this zigzag) SizeUint64(v uint64) (int) { return sizeVarint(encodeZigzag(v)) }
func (this zigzag) SizeUint32(v uint32) (int) { return this.SizeUint64(uint64(v)) }
func (this zigzag) SizeUint16(v uint16) (int) { return this.SizeUint64(uint64(v)) }
func (this zigzag) SizeUint8 (v uint8 ) (int) { return this.SizeUint64(uint64(v)) }
func (this zigzag) SizeInt64 (v  int64) (int) { return this.SizeUint64(uint64(v)) }
func (this zigzag) SizeInt32 (v  int32) (int) { return this.SizeUint64(uint64(v)) }
func (this zigzag) SizeInt16 (v  int16) (int) { return this.SizeUint64(uint64(v)) }
func (this zigzag) SizeInt8  (v  int8 ) (int) { return this.SizeUint64(uint64(v)) }
func (this zigzag) Uint64(buf []byte) (x uint64, n int) {
    u, n := Varint.Uint64(buf)
    return decodeZigzag(u), n
}
func (this zigzag) Uint32(buf []byte) (x uint32, n int) {
    u, n := this.Uint64(buf)
    if n <= 0 || u > 0xFFFFFFFF {
        return 0, 0
    }
    return uint32(u), n
}
func (this zigzag) Uint16(buf []byte) (x uint16, n int) {
    u, n := this.Uint64(buf)
    if n <= 0 || u > 0xFFFF     {
        return 0, 0
    }
    return uint16(u), n
}
func (this zigzag)  Uint8(buf []byte) (x uint8 , n int) {
    u, n := this.Uint64(buf)
    if n <= 0 || u > 0xFF       {
        return 0, 0
    }
    return uint8(u), n
}
func (this zigzag)  Int64(buf []byte) (x  int64, n int) {
    u, n := this.Uint64(buf)
    return int64(u), n
}
func (this zigzag)  Int32(buf []byte) (x  int32, n int) {
    i, n := this.Int64(buf)
    if n <= 0 || i > 0x7FFFFFFF || i <= -0x80000000 {
        return 0, 0
    }
    return int32(i), n
}
func (this zigzag)  Int16(buf []byte) (x  int16, n int) {
    i, n := this.Int64(buf)
    if n <= 0 || i > 0x7FFF     || i <= -0x8000     {
        return 0, 0
    }
    return int16(i), n
}
func (this zigzag)   Int8(buf []byte) (x  int8 , n int) {
    i, n := this.Int64(buf)
    if n <= 0 || i > 0x7F       || i <= -0x80       {
        return 0, 0
    }
    return int8 (i), n
}
func (this zigzag) PutUint64(buf []byte, x uint64) (n int) { return Varint.PutUint64(buf, encodeZigzag(x)) }
func (this zigzag) PutUint32(buf []byte, x uint32) (n int) { return this.PutUint64(buf, uint64(x)) }
func (this zigzag) PutUint16(buf []byte, x uint16) (n int) { return this.PutUint64(buf, uint64(x)) }
func (this zigzag)  PutUint8(buf []byte, x uint8 ) (n int) { return this.PutUint64(buf, uint64(x)) }
func (this zigzag)  PutInt64(buf []byte, x  int64) (n int) { return this.PutUint64(buf, uint64(x)) }
func (this zigzag)  PutInt32(buf []byte, x  int32) (n int) { return this.PutUint64(buf, uint64(x)) }
func (this zigzag)  PutInt16(buf []byte, x  int16) (n int) { return this.PutUint64(buf, uint64(x)) }
func (this zigzag)   PutInt8(buf []byte, x  int8 ) (n int) { return this.PutUint64(buf, uint64(x)) }
