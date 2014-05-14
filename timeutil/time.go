
package timeutil

import (
    "time"
    "strings"
)

const (
    YYYYMMDDHHNNSSZZZ       = "2006-01-02 15:04:05 000"
    YYYYMMDDHHNNSSZZZ_DOT   = "2006-01-02 15:04:05.000"
    ISO                     = "2006-01-02T15:04:05.000Z07:00" // RFC3339Nano
    UTC                     = "Mon, 02 Jan 2006 15:04:05 MST" // RFC1123
)

func FormatTime(t time.Time, layout string) ( s string) {
    switch layout {
    case YYYYMMDDHHNNSSZZZ_DOT, YYYYMMDDHHNNSSZZZ:
        s = t.Format(YYYYMMDDHHNNSSZZZ_DOT)
        if layout == YYYYMMDDHHNNSSZZZ {
            s = strings.Replace(s, ".", " ", -1)
        }
        return s
    default:
        return t.Format(layout)
    }
}

func ParseTime(layout, value string) (time.Time, error) {
    switch layout {
    case YYYYMMDDHHNNSSZZZ_DOT, YYYYMMDDHHNNSSZZZ:
        if !strings.Contains(value, ".") {
            parts := strings.Fields(value)
            last := len(parts)-1
            value = strings.Join(parts[:last], " ")
            value = value + "."+parts[last]
            layout = YYYYMMDDHHNNSSZZZ_DOT
        }
    }
    return time.Parse(layout, value)
}
