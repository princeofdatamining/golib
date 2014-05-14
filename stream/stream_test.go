
package stream_test

import (
    "github.com/princeofdatamining/golib/stream"
    "testing"
    "fmt"
    "io"
)

var (
    sNumbers = "0123456789"
    numbers = []byte(sNumbers)
)

func testSeekAndReadByte(t *testing.T, s *stream.Stream, off int64, whence int, want_err error, want rune) {
    var (
        err error
        msg string
        b byte
        c rune
    )
    msg = fmt.Sprintf("Seek(%d, %d),", off, whence)
    _, err = s.Seek(off, whence)
    if err != want_err {
        t.Errorf("%s want error `%v`, but got `%v`\n", msg, want_err, err)
        return
    }
    b, _ = s.ReadByte()
    c = rune(b)
    if c != want {
        t.Errorf("%s ReadByte() is `%c`, want `%c`\n", msg, c, want)
    }
}

func testReadAt(t *testing.T, s *stream.Stream, off int64, want_err error, want string) {
    var (
        err error
        msg string
        buf = []byte("......")
        str string
    )
    msg = fmt.Sprintf("ReadAt([<%d>], %d),", len(buf), off)
    _, err = s.ReadAt(buf, off)
    if err != want_err {
        t.Errorf("%s want error `%v`, but got `%v`\n", msg, want_err, err)
        return
    }
    str = string(buf)
    if str != want {
        t.Errorf("%s got `%s`, want `%s`\n", msg, str, want)
    }
}

func testUnreadByte(t *testing.T, s *stream.Stream, want_err error) {
    err := s.UnreadByte()
    if err != want_err {
        t.Errorf("UnreadByte() want error `%v`, got `%v`\n", want_err, err)
    }
}

func testPosition(t *testing.T, s *stream.Stream, want int) {
    pos := s.Position()
    if pos != want {
        t.Errorf("Position() want error `%v`, got `%v`\n", want, pos)
    }
}

func testTruncate(t *testing.T, s *stream.Stream, size int, want_err error) {
    err := s.Truncate(size)
    if err != want_err {
        t.Errorf("Truncate(%v) want error `%v`, got `%v`\n", size, want_err, err)
    }
}

func testSetSize(t *testing.T, s *stream.Stream, size int64, want_err error) {
    err := s.SetSize(size)
    if err != want_err {
        t.Errorf("SetSize(%v) want error `%v`, got `%v`\n", size, want_err, err)
    }
}

func TestStreamFrom(t *testing.T) {
    // test reader
    s := stream.NewStreamFrom(numbers)
    testSeekAndReadByte(t, s,  3, 0, nil, '3')
    testSeekAndReadByte(t, s, -2, 1, nil, '2')
    testReadAt(t, s, 7, nil, "789...")
    testSeekAndReadByte(t, s,  2, 1, nil, '5')
    testSeekAndReadByte(t, s, -2, 2, nil, '8')
    testUnreadByte(t, s, nil)
    testSeekAndReadByte(t, s,  0, 1, nil, '8')
    // test resize
    testTruncate(t, s, 20, stream.ErrNotGrowable)
    testSetSize(t, s, 20, stream.ErrNotGrowable)
    testPosition(t, s, 9)
    pos := 6
    s.Seek(int64(pos), 0)
    testTruncate(t, s, pos+1, nil)
    testPosition(t, s, pos)
    testTruncate(t, s, 0, nil)
    testPosition(t, s, 0)
    testReadAt(t, s, 5, io.EOF, "......")
    testSetSize(t, s, 10, nil)
    testPosition(t, s, 0)
    testReadAt(t, s, 5, nil, "56789.")
}

func TestStream(t *testing.T) {
}
