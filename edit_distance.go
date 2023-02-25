package badcli

import "sync/atomic"
import "unicode/utf8"

var editDistTableInUse uint32
var editDistTable []uint16

// Flag for [EditDistance]().
var NoCutoff int = -1

// Using OSA Damerau-Levenshtein distance and an optional cutoff.
// Pass [NoCutoff] if you don't want a cutoff.
//
// Panics if strings together are longer than 65535 bytes.
func EditDistance(a, b string, costCutoff int) int {
	// Idea: could adapt a version to have custom weights for different
	//       operations. Keyboard transpositions specially.

	// stupid case early return
	if costCutoff == 0 { return 0 }
	if costCutoff < 0 { costCutoff = 65535 }
	
	// safety assertion
	if len(a) + len(b) >= 65535 { panic("strings too long") }

	// determine shorter string *in bytes* (not necessarily in runes)
	short, long := a, b
	if len(long) < len(short) { short, long = long, short }

	// empty case early return
	if len(short) == 0 { return utf8.RuneCountInString(long) }

	// obtain the memoization table
	table, usingGlobTable := prepareEditDistanceTable(3*len(short) + 3) // [*]
	// [*] While 3*len(short) isn't always optimal, optimal space calculations
	//     would require runeLen() instead and I don't think that's worth it.
	
	// helper functions
	var zeroOne = func(b bool) uint16 { if b { return 1 } ; return 0 }
	var min2 = func(a, b uint16) uint16 { if a <= b { return a } ; return b }
	var min3 = func(a, b, c uint16) uint16 { return min2(min2(a, b), c) }
	
	// declare indices to work with
	row, col := 1, 1 // table row and col indices. numrows, numcols ~= runeLen(long), runeLen(short)
	iTranRow, iPrevRow, iCurrRow := 0, len(short) + 1, 2*len(short) + 2 // table indices for the 3 rows
	var prevRowRune, prevColRune rune
	costCutoff16 := uint16(costCutoff)
	if costCutoff > 65535 { costCutoff16 = 65535 }

	// pre-initialize previous row
	for i := 0; i < len(short) + 1; i++ {
		table[iPrevRow + i] = uint16(i)
	}

	// quadratic OSA algorithm
	var cost uint16
	vertStr := long
	for { // each row
		// set base row value
		minRowCost := uint16(row)
		table[iCurrRow] = minRowCost

		// obtain current row rune
		currRowRune, runeLength := utf8.DecodeRuneInString(vertStr)
		vertStr = vertStr[runeLength : ]

		horzStr := short
		for { // each column in the current row
			// obtain current column rune
			currColRune, runeLength := utf8.DecodeRuneInString(horzStr)
			horzStr = horzStr[runeLength : ]

			// core algorithm logic
			deletion     := table[iPrevRow + col + 0] + 1
			insertion    := table[iCurrRow + col - 1] + 1
			substitution := table[iPrevRow + col - 1] + zeroOne(currRowRune != currColRune)
			cost = min3(deletion, insertion, substitution)
			if row > 1 && col > 1 && currRowRune == prevColRune && prevRowRune == currColRune {
				cost = min2(cost, table[iTranRow + col - 2] + 1)
			}
			table[iCurrRow + col] = cost
			if cost < minRowCost { minRowCost = cost }

			// either stop or prepare for next iteration
			if len(horzStr) == 0 { break }
			col += 1
			prevColRune = currColRune
		}

		// either stop or prepare for next iteration
		if len(vertStr) == 0 || minRowCost >= costCutoff16 { break }
		iTranRow, iPrevRow, iCurrRow = iPrevRow, iCurrRow, iTranRow
		prevRowRune = currRowRune
		row, col = row + 1, 1
	}

	if usingGlobTable { releaseEditDistanceTable() }
	return int(min2(cost, costCutoff16))
}

func prepareEditDistanceTable(tableSize int) (table []uint16, usingGlobTable bool) {
	// acquire or make table
	if atomic.CompareAndSwapUint32(&editDistTableInUse, 0, 1) {
		// common non-concurrent case reuses a global table
		if cap(editDistTable) < tableSize {
			editDistTable = make([]uint16, tableSize)
		}
		return editDistTable[0 : tableSize], true
	} else {
		// concurrent cases will simply pay the price of creating
		// the table. this could also be better amortized, but meh
		return make([]uint16, tableSize), false
	}
}

func releaseEditDistanceTable() {
	if !atomic.CompareAndSwapUint32(&editDistTableInUse, 1, 0) {
		panic("broken code")
	}
}

