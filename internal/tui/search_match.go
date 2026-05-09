package tui

type searchMatchIndexChooser func([]int, int) int

func searchMatchIndexAtOrAfter(matchIndexes []int, currentIndex int) int {
	if len(matchIndexes) == 0 {
		return -1
	}

	for matchIndex, matchedIndex := range matchIndexes {
		if matchedIndex >= currentIndex {
			return matchIndex
		}
	}

	return 0
}

func searchMatchIndexAfter(matchIndexes []int, currentIndex int) int {
	if len(matchIndexes) == 0 {
		return -1
	}

	for matchIndex, matchedIndex := range matchIndexes {
		if matchedIndex > currentIndex {
			return matchIndex
		}
	}

	return 0
}

func searchMatchIndexBefore(matchIndexes []int, currentIndex int) int {
	if len(matchIndexes) == 0 {
		return -1
	}

	for matchIndex := len(matchIndexes) - 1; matchIndex >= 0; matchIndex-- {
		if matchIndexes[matchIndex] < currentIndex {
			return matchIndex
		}
	}

	return len(matchIndexes) - 1
}
