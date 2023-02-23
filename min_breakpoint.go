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
	iRight := len(splits)
	for iLeft := 0; iLeft < len(splits); iLeft++ {
		budget := maxLen - splits[leftIncIndices[iLeft]].leftLen
		if budget < 0 { break }

		for iRight > 0 {
			rightIndex := rightIncIndices[iRight - 1]
			if budget < splits[rightIndex].rightLen { break }
			iRight -= 1
		}

		inclusions := iLeft + (len(splits) - iRight)
		if inclusions > minBreakpointInclusions {
			minBreakpointLen = splits[leftIncIndices[iLeft]].leftLen
			minBreakpointInclusions = inclusions
		}
	}

	// return result
	return minBreakpointLen
}
