package badcli

import "testing"
import "image/color"

func TestColorStringHex(t *testing.T) {
	tests := []struct{
		in string
		out color.RGBA
	}{
		{"#000", color.RGBA{0, 0, 0, 255}},
		{"#FFF", color.RGBA{255, 255, 255, 255}},
		{"#FFF0", color.RGBA{255, 255, 255, 0}},
		{"#0000", color.RGBA{0, 0, 0, 0}},
		{"#123B", color.RGBA{17, 34, 51, 187}},
		{"#ABC", color.RGBA{170, 187, 204, 255}},
		{"#b639fa", color.RGBA{182, 57, 250, 255}},
		{"#B639FA", color.RGBA{182, 57, 250, 255}},
		{"#050038", color.RGBA{5, 0, 56, 255}},
		{"#000000FF", color.RGBA{0, 0, 0, 255}},
		{"#0a0b0cff", color.RGBA{10, 11, 12, 255}},
		{"#a011d7f0", color.RGBA{160, 17, 215, 240}},
	}

	clrString := &ColorString{}
	for i, test := range tests {
		err := clrString.ParseFromArg(test.in)
		if err != nil {
			t.Fatalf("test#%d returned an error: %s", i, err)
		}
		clrParsed := clrString.RGBA8()
		if clrParsed != test.out {
			t.Fatalf(
				"test#%d, ColorString.ParseFromArg(\"%s\") => '%v' (expected '%v')",
				i, test.in, clrParsed, test.out,
			)
		}
	}
}

func TestColorStringRGBA(t *testing.T) {
	tests := []struct{
		in string
		out color.RGBA
	}{
		{"rgb(0, 0, 0)", color.RGBA{0, 0, 0, 255}},
		{"rgb(255, 255, 255)", color.RGBA{255, 255, 255, 255}},
		{"rgb[0, 128, 128]", color.RGBA{0, 128, 128, 255}},
		{"rgb{8;13;204}", color.RGBA{8, 13, 204, 255}},
		{"rgb{100}{99}{98}", color.RGBA{100, 99, 98, 255}},
		{"RGB(100-99-98)", color.RGBA{100, 99, 98, 255}},
		{"RGB(000045-012-1)", color.RGBA{45, 12, 1, 255}},
		{"rgba(9, 9, 9, 8)", color.RGBA{9, 9, 9, 8}},
		{"rgba(10.20,30_0)", color.RGBA{10, 20, 30, 0}},
		{"RGBA( 100 , 200 , 30 , 222)", color.RGBA{100, 200, 30, 222}},
		{"RGBA( 0 1 2 111 )", color.RGBA{0, 1, 2, 111}},
	}

	clrString := &ColorString{}
	for i, test := range tests {
		err := clrString.ParseFromArg(test.in)
		if err != nil {
			t.Fatalf("test#%d returned an error: %s", i, err)
		}
		clrParsed := clrString.RGBA8()
		if clrParsed != test.out {
			t.Fatalf(
				"test#%d, ColorString.ParseFromArg(\"%s\") => '%v' (expected '%v')",
				i, test.in, clrParsed, test.out,
			)
		}
	}
}
