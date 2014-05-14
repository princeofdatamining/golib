
package uniutil

import (
    "bytes"
    "unicode/utf8"
    "encoding/binary"
)

const (
    BOM = '\uFEFF'
)

var (
    Bytes_utf8     = make([]byte, 3)
    Bytes_utf16_le = make([]byte, 2)
    Bytes_utf16_be = make([]byte, 2)
)

type BOMStyle int

const (
    BOM_None        BOMStyle = iota
    BOM_utf8
    BOM_utf16_le
    BOM_utf16_be
)

func init() () {
    n := utf8.EncodeRune(Bytes_utf8, BOM)
    Bytes_utf8 = Bytes_utf8[:n]
    //
    binary.LittleEndian.PutUint16(Bytes_utf16_le, uint16(BOM))
    binary.BigEndian   .PutUint16(Bytes_utf16_be, uint16(BOM))
}

func GetBOMBytes(style BOMStyle) ([]byte) {
    switch style {
    case BOM_utf8:
        return Bytes_utf8
    case BOM_utf16_le:
        return Bytes_utf16_le
    case BOM_utf16_be:
        return Bytes_utf16_be
    }
    return nil
}

func matchBOM(buf, bom []byte, size *int) (yes bool) {
    n := len(bom)
    if yes = len(buf) >= n && bytes.Equal(buf[:n], bom); yes {
        *size = n
    }
    return
}

func BOMTest(buf []byte, bom BOMStyle) (yes bool, n int) {
    yes = matchBOM(buf, GetBOMBytes(bom), &n)
    return
}

func BOMLen(buf []byte) (size int) {
    switch {
    case matchBOM(buf, Bytes_utf8    , &size):
    case matchBOM(buf, Bytes_utf16_le, &size):
    case matchBOM(buf, Bytes_utf16_be, &size):
    default:
    }
    return
}
