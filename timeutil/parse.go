
package timeutil

import (
    "time"
    "errors"
)


var (
    formatters = []string{
        UTC_JS,
        ISO_JS,
    }
    ErrUnsupportedFormatter = errors.New("restful: this time format is not supported")
)
func ClearTimeFormatters() () {
    formatters = nil
}
func AddTimeFormatter(formatter string) () {
    formatters = append(formatters, formatter)
}

func GuessTime(f string) (t time.Time, e error) {
    for _, formatter := range formatters {
        if t, e = time.Parse(formatter, f); e == nil {
            return 
        }
    }
    e = ErrUnsupportedFormatter
    return 
}
