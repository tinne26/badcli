package badcli

import "testing"

func TestMinBreakpointBasic(t *testing.T) {
	tests := []struct{
		in []split
		len uint16
		out uint16
	}{
		{ // test #0
			in  : []split{{1, 2}},
			len : 10,
			out : 1,
		},
		{ // test #1
			in  : []split{{1, 2}, {2, 3}},
			len : 10,
			out : 2,
		},
		{ // test #2
			in  : []split{{7, 3}, {3, 7}},
			len : 10,
			out : 3,
		},
		{ // test #3
			in  : []split{{7, 4}, {4, 7}},
			len : 10,
			out : 0,
		},
		{ // test #4
			in  : []split{{2, 5}, {3, 4}, {4, 3}, {5, 2}},
			len : 10,
			out : 5,
		},
		{ // test #5
			in  : []split{{3, 6}, {4, 5}, {5, 4}, {6, 3}},
			len : 10,
			out : 4,
		},
		{ // test #6
			in  : []split{{5, 4}, {6, 3}, {3, 6}, {4, 5}},
			len : 10,
			out : 4,
		},
		{ // test #7
			in  : []split{{8, 25}, {14, 89}, {7, 27}},
			len : 80,
			out : 8,
		},
		{ // test #8
			in  : []split{{7, 27}, {8, 25}, {14, 89}},
			len : 80,
			out : 8,
		},
	}

	for i, test := range tests {
		result := findBreakpointMin(test.in, test.len)
		if result != test.out {
			t.Fatalf("test#%d, findMinBreakpoint(%v, %d) = %d (expected %d)", i, test.in, test.len, result, test.out)
		}
	}
}
