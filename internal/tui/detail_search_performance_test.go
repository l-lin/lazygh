package tui

import (
	"strings"
	"testing"
)

func TestDetailViewStateSyncSearch_GivenTheSameDocumentAndQuery_WhenCalledRepeatedly_ThenItReusesCachedMatchesWithoutAllocations(t *testing.T) {
	document := newDetailDocument(strings.Repeat("Alpha beta gamma\n", 200), 80)
	subject := newDetailViewState()
	subject.cursor = detailPosition{line: 0, column: 0}

	subject.syncSearch(document, "Alpha")
	if len(subject.searchMatches) == 0 {
		t.Fatal("expected the initial search to find matches")
	}

	actual := testing.AllocsPerRun(100, func() {
		subject.syncSearch(document, "Alpha")
	})

	if actual != 0 {
		t.Fatalf("expected repeated search sync allocations %v, actual %v", 0.0, actual)
	}
}
