package badcli

import "sort"

type split struct {
	leftLen  uint16
	rightLen uint16
}

func findBreakpointMin(splits []split, maxLen uint16) uint16 {
	// empty case
	if len(splits) == 0 { return 0 }
	
	// create index slice
	indices := make([]int, len(splits))
	for i := 0; i < len(splits); i++ { indices[i] = i }
	
	// sort indices slice by increasing left length
	sort.Slice(indices, func(i, j int) bool {
		return splits[indices[i]].leftLen < splits[indices[j]].leftLen
	})

	// brute force search
	minBreakpointLen        := uint16(0)
	minBreakpointInclusions := 0
	for index := 0; index < len(splits); index++ {
		split := splits[indices[index]]
		if split.leftLen + split.rightLen > maxLen { continue }
		if index < minBreakpointInclusions { break }

		rightBudget := maxLen - split.leftLen
		inclusions := 1
		for subIndex := 0; subIndex < index; subIndex++ {
			subSplit := splits[indices[subIndex]]
			if subSplit.rightLen <= rightBudget {
				inclusions += 1
			}
		}

		if inclusions > minBreakpointInclusions {
			minBreakpointLen = split.leftLen
			minBreakpointInclusions = inclusions
		}		
	}

	// return result
	return minBreakpointLen
}
