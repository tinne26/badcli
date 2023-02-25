package badcli

import "sort"

type split struct {
	leftLen  uint16
	rightLen uint16
}

func findBreakpointMin(splits []split, maxLen uint16) uint16 {
	// empty case
	if len(splits) == 0 { return 0 }
	
	// create left and right index slices in increasing length
	leftIncIndices := make([]int, len(splits))
	for i := 0; i < len(splits); i++ { leftIncIndices[i] = i }
	rightIncIndices := make([]int, len(splits))
	copy(rightIncIndices, leftIncIndices)
	
	// sort both slices by increasing length
	sort.Slice(leftIncIndices, func(i, j int) bool {
		return splits[leftIncIndices[i]].leftLen < splits[leftIncIndices[i]].leftLen
	})
	sort.Slice(rightIncIndices, func(i, j int) bool {
		return splits[rightIncIndices[i]].leftLen < splits[rightIncIndices[i]].leftLen
	})

	// linear search
	minBreakpointLen        := uint16(0)
	minBreakpointInclusions := 0
	iRight := len(splits) - 1

	for iLeft := 0; iLeft < len(splits); iLeft++ {
		leftCost := splits[leftIncIndices[iLeft]].leftLen
		if leftCost > maxLen { break }
		budget := maxLen - leftCost

		accept := false
		for {
			rightIndex := rightIncIndices[iRight]
			if budget < splits[rightIndex].rightLen { break }
			accept = true
			if iRight == 0 { break } // can't go further
			iRight -= 1
		}
		if !accept { continue }

		inclusions := iLeft + (len(splits) - iRight)
		if inclusions > minBreakpointInclusions {
			minBreakpointLen = leftCost
			minBreakpointInclusions = inclusions
		}		
	}

	// return result
	return minBreakpointLen
}
