package badcli

import "errors"
import "strconv"
import "strings"
import "unicode"
import "unicode/utf8"
import "image/color"

var ColorStringFormatsInfo =
	"- Hexadecimal: \"#234\", \"#234F\", \"#332270\", \"#A8B9CAF0\".\n" +
	"- Explicit RGB triplets: \"rgb(9, 44, 128)\", \"RGB(4, 9, 5)\", \"rgb(12;0;24)\". The values fall " + 
	"between 0 and 255.\n" +
	"- Explicit RGBA quadruplets: like RGB, but using the \"rgba\" or \"RGBA\" descriptor and four " +
	"values instead of three (e.g. \"rgba(0, 255, 0, 128)\").\n" +
	"- Implicit RGB(A): triplets or quadruplets for RGB(A) can also be passed without the descriptor, " +
	"without braces, and with any punctuation symbol as a separator in general. This means that " +
	"\"[255;0;128]\", \"80, 90, 60\", \"{0, 0, 255, 128}\", \"(99:98:97)\" and \"100.100.200.200\" are " +
	"all weird but allowed."

// Assert interface compliance.
var _ FlagValue = (*ColorString)(nil)

type ColorString color.RGBA

func NewColorString(r, g, b uint8) *ColorString {
	aux := ColorString(color.RGBA{r, g, b, 255})
	return &aux
}

func (self *ColorString) RGBA8() color.RGBA {
	return color.RGBA(*self)
}

func (self *ColorString) SetRGBA8(clr color.RGBA) {
	*self = ColorString(clr)
}

func (self *ColorString) String() string {
	clr := color.RGBA(*self)
	rStr := strconv.Itoa(int(clr.R))
	gStr := strconv.Itoa(int(clr.G))
	bStr := strconv.Itoa(int(clr.B))
	if clr.A == 255 {
		return "rgb(" + rStr + ", " + gStr + ", " + bStr + ")"
	} else {
		aStr := strconv.Itoa(int(clr.A))
		return "rgba(" + rStr + ", " + gStr + ", " + bStr + ", " + aStr + ")"
	}
}

func (self *ColorString) ParseFromArg(arg string) error {
	// cleanup
	arg = strings.TrimSpace(arg)
	
	// hexadecimal notation case
	if strings.HasPrefix(arg, "#") {
		switch len(arg) {
		case 4: // #RGB
			var channels [3]uint8
			for i := 0; i < 3; i += 1 {
				char := arg[i + 1]
				hex, err := hexCharValue(char)
				if err != nil { return err }
				channels[i] = (hex << 4) + hex
			}
			*self = ColorString(color.RGBA{channels[0], channels[1], channels[2], 255})
		case 5: // #RGBA
			var channels [4]uint8
			for i := 0; i < 4; i += 1 {
				char := arg[i + 1]
				hex, err := hexCharValue(char)
				if err != nil { return err }
				channels[i] = (hex << 4) + hex
			}
			*self = ColorString(color.RGBA{channels[0], channels[1], channels[2], channels[3]})
		case 7: // #RRGGBB
			var channels [3]uint8
			for i := 0; i < 3; i += 1 {
				highChar, lowChar := arg[i*2 + 1], arg[i*2 + 2]
				high, err := hexCharValue(highChar)
				if err != nil { return err }
				low , err := hexCharValue(lowChar)
				if err != nil { return err }
				channels[i] = (high << 4) + low
			}
			*self = ColorString(color.RGBA{channels[0], channels[1], channels[2], 255})
		case 9: // #RRGGBBAA
			var channels [4]uint8
			for i := 0; i < 4; i += 1 {
				highChar, lowChar := arg[i*2 + 1], arg[i*2 + 2]
				high, err := hexCharValue(highChar)
				if err != nil { return err }
				low , err := hexCharValue(lowChar)
				if err != nil { return err }
				channels[i] = (high << 4) + low
			}
			*self = ColorString(color.RGBA{channels[0], channels[1], channels[2], channels[3]})
		default:
			return errors.New("invalid hexadecimal color encoding (invalid length)")
		}

		return nil
	}

	// standard rgb/rgba notation case
	needsBraces := false
	minComponents := 3
	maxComponents := 4
	if strings.HasPrefix(arg, "rgb") || strings.HasPrefix(arg, "RGB") {
		needsBraces = true
		if strings.HasPrefix(arg, "rgba") || strings.HasPrefix(arg, "RGBA") {
			arg = strings.TrimSpace(arg[4 : ])
			minComponents = 4
		} else {
			arg = strings.TrimSpace(arg[3 : ])
			maxComponents = 3
		}
	}

	hasBraces := false
	for _, bracesPair := range []struct{ Left, Right rune }{{'(', ')'}, {'[', ']'}, {'{', '}'}, {'<', '>'}, {'«', '»'}} {
		leftRune, leftLen := utf8.DecodeRuneInString(arg)
		if leftRune == bracesPair.Left {
			rightRune, rightLen := utf8.DecodeLastRuneInString(arg)
			if rightRune == bracesPair.Right {
				hasBraces = true
				arg = strings.TrimSpace(arg[leftLen : len(arg) - rightLen])
				break
			}
		}
	}

	// multiple error checks
	if needsBraces && !hasBraces {
		return errors.New("expected symmetric braces surrounding the color channel values (e.g. \"rgb(200, 128, 0)\")")
	}
	if len(arg) < minComponents*2 - 1 {
		return errors.New("incomplete color definition")
	}
	if arg[0] < 48 || arg[0] > 57 {
		return errors.New("invalid color format: expected a number, braces or a prefix (e.g. \"rgb\", \"#\")")
	}

	// read digits
	var components [4]int
	componentIndex := 0
	readingSeparator := true
	for len(arg) > 0 {
		runeChar, runeLen := utf8.DecodeRuneInString(arg)
		if runeChar >= 48 && runeChar <= 57 { // valid digit
			if readingSeparator { componentIndex += 1 }
			readingSeparator = false
			if componentIndex > maxComponents {
				return errors.New("too many color components (expected only " + strconv.Itoa(maxComponents) + ")")
			}
			component := components[componentIndex - 1]
			newComponentValue := component*10 + int(byte(runeChar) - 48)
			if newComponentValue > 255 {
				return errors.New("color component can't exceed 255")
			}
			components[componentIndex - 1] = newComponentValue
		} else {
			readingSeparator = true
			if runeChar != ' ' && !unicode.IsPunct(runeChar) && runeChar != '|' {
				charStr := string(runeChar)
				if runeChar < 128 { charStr = asciiCodePointName(byte(runeChar)) }
				return errors.New("invalid color component separator '" + charStr + "' (only spaces and punctuation symbols allowed)")
			}
		}
		arg = arg[runeLen : ]
	}

	if componentIndex < minComponents {
		if minComponents < maxComponents {
			return errors.New("too few color components (expected at least " + strconv.Itoa(minComponents) + ")")
		} else {
			return errors.New("too few color components (expected " + strconv.Itoa(minComponents) + " components)")
		}
	}

	if componentIndex == 4 {
		*self = ColorString(color.RGBA{
			uint8(components[0]), 
			uint8(components[1]), 
			uint8(components[2]), 
			uint8(components[3]),
		})
	} else {
		*self = ColorString(color.RGBA{
			uint8(components[0]),
			uint8(components[1]),
			uint8(components[2]),
			255,
		})
	}
	return nil
}

// ---- helpers ----

var CtrlAsciiNames = []string{
	"NUL", "SOH", "STX", "ETX", "EOT", "ENQ", "ACK", "BEL",
	"BS" , "TAB", "LF" , "VT" , "FF" , "CR" , "SO" , "SI" ,
	"DLE", "DC1", "DC2", "DC3", "DC4", "NAK", "SYN", "ETB",
	"CAN", "EM" , "SUB", "ESC", "FS" , "GS" , "RS" , "US",
}

func asciiCodePointName(char byte) string {
	if char < 32 { return CtrlAsciiNames[char] }
	if char == 127 { return "DEL" }
	return string(char)
}

func hexCharValue(char byte) (uint8, error) {
	var err = func(b byte) error {
		msg := "invalid hexadecimal character '" + asciiCodePointName(b)
		msg += "' (ascii code = " + strconv.Itoa(int(b)) +")"
		return errors.New(msg)
	}

	if char < 48  { return 0, err(char) }
	if char < 58  { return char - 48, nil }
	if char < 65  { return 0, err(char) }
	if char < 71  { return char - 65 + 10, nil }
	if char < 97  { return 0, err(char) }
	if char < 103 { return char - 97 + 10, nil }
	return 0, err(char)
}
