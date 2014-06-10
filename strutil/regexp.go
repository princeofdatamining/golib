
package strutil

import (
)

type SubmatchIndex []int

func (this SubmatchIndex) GetIndexPairs(no int) (start, end int, valid bool) {
    start, end = this[no*2], this[no*2+1]
    valid = start >= 0 && end >= 0
    return 
}
func (this SubmatchIndex) GetSubmatch(s string, no int) (sub string, start, end int, valid bool) {
    if start, end, valid = this.GetIndexPairs(no); valid {
        sub = s[start:end]
    }
    return 
}
func (this SubmatchIndex) GetSubstring(s string, no int) (sub string) {
    sub, _, _, _ = this.GetSubmatch(s, no)
    return 
}
