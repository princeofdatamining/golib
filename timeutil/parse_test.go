
package timeutil_test

import (
    "testing"
    "time"
    "github.com/princeofdatamining/golib/timeutil"
)

var (
    tmAliOss = "2014-09-03 05:48:35 +0000"
    tmISO_JS = "2014-09-03T08:37:56.249Z"
    tmUTC_JS = "Wed, 03 Sep 2014 08:37:56 GMT"
)
type guessResult struct {
    text    string
    err     error
    value   time.Time
}
type guessCase struct {
    formatter   string
    results     []guessResult
}
var guessTimeCases = []guessCase{
    {
        formatter: "",
        results: []guessResult{
            {
                text: tmAliOss,
                err: timeutil.ErrUnsupportedFormatter,
            },
            {
                text: tmISO_JS,
                err: timeutil.ErrUnsupportedFormatter,
            },
            {
                text: tmUTC_JS,
                err: timeutil.ErrUnsupportedFormatter,
            },
        },
    },
    {
        formatter: "2006-01-02 15:04:05 -0700",
        results: []guessResult{
            {
                text: tmAliOss,
            },
            {
                text: tmISO_JS,
                err: timeutil.ErrUnsupportedFormatter,
            },
            {
                text: tmUTC_JS,
                err: timeutil.ErrUnsupportedFormatter,
            },
        },
    },
    {
        formatter: timeutil.UTC_JS,
        results: []guessResult{
            {
                text: tmAliOss,
            },
            {
                text: tmISO_JS,
                err: timeutil.ErrUnsupportedFormatter,
            },
            {
                text: tmUTC_JS,
            },
        },
    },
    {
        formatter: timeutil.ISO_JS,
        results: []guessResult{
            {
                text: tmAliOss,
            },
            {
                text: tmISO_JS,
            },
            {
                text: tmUTC_JS,
            },
        },
    },
}
func TestGuessTime(t *testing.T) () {
    timeutil.ClearTimeFormatters()
    for step, kase := range guessTimeCases {
        if kase.formatter != "" {
            timeutil.AddTimeFormatter(kase.formatter)
        }
        for _, result := range kase.results {
            _, e := timeutil.GuessTime(result.text)
            if e != result.err {
                t.Errorf("[Step %d] Guess(%q) want error %v, but got %v\n", step, result.text, result.err, e)
            }
        }
    }
}
