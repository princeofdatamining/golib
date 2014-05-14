
package stream

import (
    "errors"
    "io"
    "bufio"
    "unicode/utf8"
)

const (
    defaultStreamSize = 4096
)

var (
    ErrInvalidWhence    = errors.New("stream: invalid whence")
    ErrInvalidOffset    = errors.New("stream: invalid offset")
    ErrNegativePosition = errors.New("stream: negative position")
    ErrUnreadByte       = errors.New("stream: can't UnreadByte at beginning of slice")
    ErrUnreadRune       = errors.New("stream: previous operation was not ReadRune")
    ErrNotGrowable      = errors.New("stream: can't operate a ungrowable buffer")
)

func newStreamSize(b []byte, n int, growable bool, align bool) (*Stream) {
    return &Stream{
        raw: b,
        capacity: int64(len(b)),
        buf: b[:n],
        size: int64(n),
        growable: growable,
        memAlign: align,
        prevRune: -1,
    }
}

func newStream(b []byte, growable bool, align bool) (*Stream) {
    return newStreamSize(b, len(b), growable, align)
}

func NewStreamFrom(b []byte) (*Stream) {
    return newStream(b, false, false)
}

func NewStreamFixed(size int) (*Stream) {
    return newStream(make([]byte, size), false, false)
}

func NewStreamAlign(size int, align bool) (*Stream) {
    if size < 0 {
        size = 0
    }
    capacity := size
    if capacity <= 0 {
        capacity = defaultStreamSize
    }
    return newStreamSize(make([]byte, capacity), size, true, align)
}

func NewStream(size int) (*Stream) {
    return NewStreamAlign(size, true)
}

type Stream struct {
    raw         []byte
    buf         []byte
    capacity    int64
    growable    bool
    memAlign    bool
    pos         int64
    size        int64
    prevRune    int64
}

func (this *Stream) Raw() ([]byte) {
    return this.raw
}

func (this *Stream) Buf() ([]byte) {
    return this.buf
}

func (this *Stream) Size64() (int64) {
    return this.size
}

func (this *Stream) Size() (int) {
    return int(this.size)
}

func (this *Stream) Capacity64() (int64) {
    return this.capacity
}

func (this *Stream) Capacity() (int) {
    return int(this.capacity)
}

func (this *Stream) Position() (int) {
    return int(this.pos)
}

func (this *Stream) Len() (n int) {
    n = int(this.size - this.pos)
    return
}

func (this *Stream) Bytes() ([]byte) {
    return this.buf[this.pos:]
}

func (this *Stream) String() (string) {
    if this == nil {
        return ""
    }
    return string(this.buf[this.pos:])
}

func (this *Stream) Seek(off int64, whence int) (pos int64, err error) {
    this.prevRune = -1
    switch whence {
    case 0:
        pos = off
    case 1:
        pos = off + this.pos
    case 2:
        pos = off + this.size
    default:
        return 0, ErrInvalidWhence
    }
    if pos < 0 {
        return 0, ErrNegativePosition
    }
    this.pos = pos
    return
}

func (this *Stream) Next(count int) (data []byte, err error) {
    this.prevRune = -1
    n := int64(count)
    if n < 0 {
        return nil, bufio.ErrNegativeCount
    }
    if this.ensureSize(this.pos+n) != nil {
        n = this.size - this.pos
        err = bufio.ErrBufferFull
    }
    data = this.buf[this.pos:this.pos+n]
    this.pos += n
    return data, err
}

// ** Reader methods

func (this *Stream) Read(b []byte) (n int, err error) {
    this.prevRune = -1
    if len(b) == 0 {
        return 0, nil
    }
    if this.pos >= this.size {
        return 0, io.EOF
    }
    n = copy(b, this.buf[this.pos:])
    this.pos += int64(n)
    return
}

func (this *Stream) ReadAt(b []byte, off int64) (n int, err error) {
    if len(b) == 0 {
        return 0, nil
    }
    if off < 0 {
        return 0, ErrInvalidOffset
    }
    if off >= this.size {
        return 0, io.EOF
    }
    n = copy(b, this.buf[int(off):])
    return n, nil
}

func (this *Stream) ReadByte() (b byte, err error) {
    this.prevRune = -1
    if this.pos >= this.size {
        return 0, io.EOF
    }
    b = this.buf[this.pos]
    this.pos++
    return b, nil
}

func (this *Stream) UnreadByte() (err error) {
    this.prevRune = -1
    if (this.pos <= 0) {
        return ErrUnreadByte
    }
    this.pos--
    return nil
}

func (this *Stream) ReadRune() (ch rune, size int, err error) {
    if this.pos >= this.size {
        return 0, 0, io.EOF
    }
    this.prevRune = this.pos
    if c := this.buf[this.pos]; c < utf8.RuneSelf {
        this.pos++
        return rune(c), 1, nil
    }
    ch, size = utf8.DecodeRune(this.buf[this.pos:])
    this.pos += int64(size)
    return ch, size, nil
}

func (this *Stream) UnreadRune() (err error) {
    if this.prevRune < 0 {
        return ErrUnreadRune
    }
    this.pos = this.prevRune
    this.prevRune = -1
    return nil
}

func (this *Stream) WriteTo(w io.Writer) (n int64, err error) {
    this.prevRune = -1
    if this.pos >= this.size {
        return 0, nil
    }
    b := this.buf[this.pos:]
    var size int
    size, err = w.Write(b)
    n = int64(size)
    this.pos += n
    if size != len(b) && err == nil {
        err = io.ErrShortWrite
    }
    return
}

// ** resize methods

func (this *Stream) Reset() (err error) {
    return this.resize(0, false)
}

func (this *Stream) Truncate(n int) (err error) {
    return this.resize(int64(n), false)
}

func (this *Stream) SetSize(size int64) (err error) {
    return this.resize(size, this.growable)
}

func (this *Stream) resize(size int64, growable bool) (err error) {
    if size < 0 {
        size = 0
    }
    if (size > this.capacity) && !growable {
        return ErrNotGrowable
    }
    if this.pos > size {
        this.pos = size
        this.prevRune = -1
    }
    if this.prevRune > size {
        this.prevRune = -1
    }
    if size > this.capacity {
        capacity := size
        if this.memAlign {
            capacity = (size+defaultStreamSize-1) & ^(defaultStreamSize-1)
        }
        dummy := make([]byte, capacity)
        copy(dummy, this.raw)
        this.raw = dummy
        this.capacity = capacity
    }
    this.buf = this.raw[:size]
    this.size = size
    return nil
}

func (this *Stream) ensureSize(need int64) (err error) {
    if this.size < need {
        err = this.SetSize(need)
    }
    return
}

// ** Writer methods

func (this *Stream) Write(b []byte) (n int, err error) {
    this.prevRune = -1
    buflen := len(b)
    if buflen == 0 {
        return 0, nil
    }
    err = this.ensureSize(this.pos+int64(buflen))
    n = copy(this.buf[this.pos:], b)
    this.pos += int64(n)
    if n < buflen {
        err = io.ErrShortWrite
    }
    return
}

func (this *Stream) WriteAt(b []byte, off int64) (n int, err error) {
    buflen := len(b)
    if buflen == 0 {
        return 0, nil
    }
    if off < 0 {
        return 0, ErrInvalidOffset
    }
    err = this.ensureSize(off+int64(buflen))
    n = copy(this.buf[int(off):], b)
    if n < buflen {
        err = io.ErrShortWrite
    }
    return n, err
}

func (this *Stream) WriteByte(b byte) (err error) {
    this.prevRune = -1
    err = this.ensureSize(this.pos+1)
    if err == nil {
        this.buf[this.pos] = b
        this.pos++
    } else {
        err = io.ErrShortWrite
    }
    return
}

func (this *Stream) WriteString(s string) (n int, err error) {
    return this.Write([]byte(s))
}

func (this *Stream) ReadFrom(r io.Reader) (n int64, err error) {
    this.prevRune = -1
    var m int
    for {
        // keep buffer not full if it's growable
        this.ensureSize(this.pos+1)
        m, err = r.Read(this.buf[this.pos:])
        if m == 0 {
            break
        }
        this.pos += int64(m)
        n += int64(m)
        if err != nil {
            break
        }
    }
    if err == io.EOF {
        err = nil
    }
    return
}

// ** Other methods
