package badcli

import "strconv"
import "unicode/utf8"

// Given a paragraph and a maximum line length, calls lineFunc for each
// line split from the paragraph. The line splitting algorithm is extremely
// basic, you should look into the Knuth and Plass implementation for TeX
// or Android's Minikin if you need something decent instead.
// TODO: I most definitely want to support "- " at start of line
// TODO: yes this is extremely broken, I have to actually fix everything.
func EachLine(paragraph string, maxLen int, lineFunc func(string) error) error {
	lineStart := 0
	lineEnd   := 0
	lineRuneCount := 0
	index := 0

	var flushLine = func(start int) error {
		err := lineFunc(paragraph[lineStart : lineEnd])
		index = start
		lineStart = start
		lineRuneCount = 0
		return err
	}

	for index < len(paragraph) {
		fragment := getNextFragment(paragraph, index)
retry:
		if fragment.IsLineBreak {
			err := flushLine(fragment.EndIndex)
			if err != nil { return err }
		} else {
			if fragment.RuneLength + lineRuneCount < maxLen {
				// fragment fits in current line
				lineRuneCount += fragment.RuneLength 
				if !fragment.CanOmitAtEnd {
					lineEnd = fragment.EndIndex
				}
				index = fragment.EndIndex
			} else { // fragment doesn't fit in current line
				if fragment.RuneLength > maxLen/3 && lineRuneCount <= maxLen/2 {
					// force part of the long fragment to be pushed to line anyway
					lineEnd = fragment.StartIndex
					fragment.RuneLength = maxLen - lineRuneCount
					for i := fragment.RuneLength; i > 0; i-- {
						_, runeSize := utf8.DecodeRuneInString(paragraph[lineEnd : ])
						lineEnd += runeSize
					}
					err := flushLine(lineEnd)
					if err != nil { return err }
					fragment.StartIndex = lineEnd
					lineEnd = fragment.EndIndex
				} else {
					// flush current line and jump to next
					err := flushLine(fragment.EndIndex)
					if err != nil { return err }
				}
				goto retry
			}
		}
	}

	if lineStart < index {
		lineFunc(paragraph[lineStart : lineEnd])
	}

	return nil
}

type fragment struct {
	StartIndex int
	EndIndex int
	RuneLength int
	CanOmitAtEnd bool
	IsLineBreak bool // special flag
}

func getNextFragment(paragraph string, index int) *fragment {
	if index >= len(paragraph) {
		return &fragment{ IsLineBreak: true }
	}

	frag := &fragment{ StartIndex: index, EndIndex: index }
	for frag.EndIndex < len(paragraph) {
		runeChar, runeSize := utf8.DecodeRuneInString(paragraph[frag.EndIndex : ])
		switch runeChar {
		case '\n':
			if frag.RuneLength == 0 || frag.CanOmitAtEnd {
				frag.IsLineBreak = true
				frag.EndIndex += runeSize
			}
			return frag
		case ' ':
			if frag.RuneLength == 0 || frag.CanOmitAtEnd {
				frag.CanOmitAtEnd = true
				frag.EndIndex += runeSize
				frag.RuneLength += 1
			} else {
				return frag
			}
		case '-':
			frag.EndIndex += runeSize
			frag.RuneLength += 1
			return frag
		default:
			// panic on control chars
			if runeChar < 32 || runeChar == 127 {
				// NOTICE: no good reason to support \r. For \t, it could be considered, but since
				//         the tabs are configurable, I can't really say how much space it will take.
				//         I could convert it to 4 spaces or something, or consider it a single space
				//         in terms of processing, but I don't like any of that. I don't know?
				panic("unexpected control character " + strconv.Itoa(int(runeChar)) + " in paragraph")
			}

			// and then do the normal stuff
			if frag.CanOmitAtEnd {
				// we were handling spaces, so we have to stop here
				return frag
			}

			// otherwise add the char to the list
			frag.EndIndex += runeSize
			frag.RuneLength += 1
		}
	}

	return frag
}
