
package timeutil_test

import (
    "time"
    "testing"
    "github.com/princeofdatamining/golib/timeutil"
)

type inTimeData struct {
    t       time.Time
    layout  string
    s       string
}
var (
    t1 = time.Date(2013, time.Month(4), 30, 14, 22, 30, 10*int(time.Millisecond), time.UTC)
    t2 = t1.Add(-time.Duration(t1.Nanosecond()))
)
var inTimeTests = []*inTimeData{
    { t1, timeutil.YYYYMMDDHHNNSSZZZ_DOT, "2013-04-30 14:22:30.010", },
    { t1, timeutil.YYYYMMDDHHNNSSZZZ    , "2013-04-30 14:22:30 010", },
    { t1, timeutil.ISO_JS               , "2013-04-30T14:22:30.010Z", },
    { t2, timeutil.UTC_JS               , "Tue, 30 Apr 2013 14:22:30 UTC", },
}
func testParseAndFormat(t *testing.T, in *inTimeData) {
    s := timeutil.FormatTime(in.t, in.layout)
    if s != in.s {
        t.Fatalf("format %s error, got %s\n", in.s, s)
    }
    tm, err := timeutil.ParseTime(in.layout, in.s)
    if err != nil {
        t.Fatalf("parse  %s error: %v\n", in.s, err)
    }
    if tm != in.t {
        t.Fatalf("parse  %s error, got %v\n", in.s, tm)
    }
}
func TestParseAndFormat(t *testing.T) {
    for _, in := range inTimeTests {
        testParseAndFormat(t, in)
    }
}
