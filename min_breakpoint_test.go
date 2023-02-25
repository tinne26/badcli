package badcli

import "testing"

func TestMinBreakpointBasic(t *testing.T) {
	tests10 := []struct{
		in []split
		out uint16
	}{
		{ // test #0
			in  : []split{{1, 2}},
			out : 1,
		},
		{ // test #1
			in  : []split{{1, 2}, {2, 3}},
			out : 2,
		},
		{ // test #2
			in  : []split{{7, 3}, {3, 7}},
			out : 3,
		},
		{ // test #3
			in  : []split{{7, 4}, {4, 7}},
			out : 0,
		},
		{ // test #4
			in  : []split{{2, 5}, {3, 4}, {4, 3}, {5, 2}},
			out : 5,
		},
		{ // test #5
			in  : []split{{3, 6}, {4, 5}, {5, 4}, {6, 3}},
			out : 4,
		},
		{ // test #6
			in  : []split{{5, 4}, {6, 3}, {3, 6}, {4, 5}},
			out : 4,
		},
	}

	maxLen := uint16(10)
	for i, test := range tests10 {
		result := findBreakpointMin(test.in, maxLen)
		if result != test.out {
			t.Fatalf("test#%d, findMinBreakpoint(%v, %d) = %d (expected %d)", i, test.in, maxLen, result, test.out)
		}
	}
}
