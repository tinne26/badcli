package badcli

import "testing"

func TestEditDistanceBasic(t *testing.T) {
	tests := []struct{
		in1 string
		in2 string
		out int
	}{ // tests mostly stolen from hbollon/go-edlib
		{"a", "a", 0},
		{"", "abcde", 5},
		{"abcde", "", 5},
		{"abcde", "abcde", 0},
		{"ab", "aa", 1},
		{"ca", "abc", 3},
		{"abc", "aac", 1},
		{"aabc", "aaac", 1},
		{"ab", "ba", 1},
		{"ab", "aaa", 2},
		{"bbb", "a", 3},
		{"a cat", "an abct", 4},
		{"tears for fears", "fears for tears", 2},
		{"jellyfish", "smellyfish", 2},
		{"ó_ò", "ò_ó", 2},
		{"こにんち", "こんにちは", 2},
		{"🙂😄🙂😄", "😄🙂😄🙂", 2},
	}

	for i, test := range tests {
		dist := EditDistance(test.in1, test.in2, NoCutoff)
		if dist != test.out {
			t.Fatalf(
				"test#%d, EditDistance(\"%s\", \"%s\", NoCutoff) = %d (expected %d)",
				i, test.in1, test.in2, dist, test.out,
			)
		}
	}
}

func TestEditDistanceCutoff(t *testing.T) {
	tests := []struct{
		in1 string
		in2 string
		inCutoff int
		out int
	}{
		{"a", "b", 0, 0},
		{"magnificient", "mafgnicient", 3, 3},
		{"long-disaster", "hello-world", 15, 13},
		{"long-disaster", "hello-world", 13, 13},
		{"long-disaster", "hello-world", 12, 12},
		{"long-disaster", "hello-world", 6, 6},
		{"tears for fears", "fears for tears", 5, 2},
		{"tears for fears", "fears for tears", 1, 1},
		{"tears for fears", "fears for tears", 2, 2},
		{"tears for fears", "sraef rof sraet", 50, 10},
		{"tears for fears", "sraef rof sraet", 5, 5},
		{"tears for fears", "sraef rof sraet", 10, 10},
		{"tears for fears", "sraef rof sraet", 1, 1},
	}

	for i, test := range tests {
		dist := EditDistance(test.in1, test.in2, test.inCutoff)
		if dist != test.out {
			t.Fatalf(
				"test#%d, EditDistance(\"%s\", \"%s\", %d) = %d (expected %d)",
				i, test.in1, test.in2, test.inCutoff, dist, test.out,
			)
		}
	}
}
