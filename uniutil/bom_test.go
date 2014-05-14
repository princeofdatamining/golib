
package uniutil

import (
    "testing"
)

// test BOMLen

type len_sample struct {
    IN          []byte
    Expected    []int
    Text        string
}
var len_samples = []*len_sample{
    { Bytes_utf8    , []int{ 0, 0, 0, 3,    }, "bom_utf8"     },
    { Bytes_utf16_le, []int{ 0, 0, 2,       }, "bom_utf16_le" },
    { Bytes_utf16_be, []int{ 0, 0, 2,       }, "bom_utf16_be" },
}
func testBOMLen(t *testing.T, in *len_sample) {
    for c := 0; c <= len(in.IN); c++ {
        expect := in.Expected[c]
        if expect < 0 {
            continue
        }
        n := BOMLen(in.IN[:c])
        if expect != n {
            t.Errorf("BOMLen(%s[:%d]) = %d, but got %d\n", in.Text, c, expect, n)
        }
    }
}
func TestBOMLen(t *testing.T) {
    for _, sample := range len_samples {
        testBOMLen(t, sample)
    }
}

// test BOMTest

type check_sample struct {
    IN          []byte
    Expected    map[BOMStyle][]int
    Text        string
}
var check_samples = []*check_sample{
    {
        IN: Bytes_utf8    ,
        Text: "bom_utf8"    ,
        Expected: map[BOMStyle][]int{
            BOM_None    : []int{ 0, 0, 0, 0 },
            BOM_utf8    : []int{ 0, 0, 0, 3 },
            BOM_utf16_le: []int{ 0, 0, 0, 0 },
            BOM_utf16_be: []int{ 0, 0, 0, 0 },
        },
    },
    {
        IN: Bytes_utf16_le,
        Text: "bom_utf16_le",
        Expected: map[BOMStyle][]int{
            BOM_None    : []int{ 0, 0, 0 },
            BOM_utf8    : []int{ 0, 0, 0 },
            BOM_utf16_le: []int{ 0, 0, 2 },
            BOM_utf16_be: []int{ 0, 0, 0 },
        },
    },
    {
        IN: Bytes_utf16_be,
        Text: "bom_utf16_be",
        Expected: map[BOMStyle][]int{
            BOM_None    : []int{ 0, 0, 0 },
            BOM_utf8    : []int{ 0, 0, 0 },
            BOM_utf16_le: []int{ 0, 0, 0 },
            BOM_utf16_be: []int{ 0, 0, 2 },
        },
    },
}
func testBOMCheck(t *testing.T, in *check_sample) {
    for c := 0; c <= len(in.IN); c++ {
        buf := in.IN[:c]
        for style, expects := range in.Expected {
            _, n := BOMTest(buf, style)
            expect := expects[c]
            if expect != n {
                t.Errorf("BOMTest(%s[:%d], %v) must be %d, bug got %d\n", in.Text, c, style, expect, n)
            }
        }
    }
}
func TestBOMCheck(t *testing.T) {
    for _, sample := range check_samples {
        testBOMCheck(t, sample)
    }
}
