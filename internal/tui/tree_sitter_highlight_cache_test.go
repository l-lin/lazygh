package tui

import (
	"reflect"
	"testing"
)

func TestTreeSitterSyntaxRanges_GivenCachedRanges_WhenMutatingTheReturnedSlice_ThenLaterReadsKeepTheOriginalValue(t *testing.T) {
	subject := newSyntaxRangeCache(2)
	subject.Put("alpha", []styledRuneRange{{start: 1, end: 3, prefix: "prefix-a"}})

	actual, actualOK := subject.Get("alpha")
	if !actualOK {
		t.Fatal("expected cached ranges for alpha")
	}
	actual[0].prefix = "mutated"

	actual, actualOK = subject.Get("alpha")
	if !actualOK {
		t.Fatal("expected cached ranges for alpha on the second read")
	}

	expected := []styledRuneRange{{start: 1, end: 3, prefix: "prefix-a"}}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected cached ranges %v, actual %v", expected, actual)
	}
}

func TestTreeSitterSyntaxRanges_GivenCacheCapacityExceeded_WhenAddingANewEntry_ThenItEvictsTheLeastRecentlyUsedValue(t *testing.T) {
	subject := newSyntaxRangeCache(2)
	subject.Put("alpha", []styledRuneRange{{start: 1, end: 2, prefix: "prefix-a"}})
	subject.Put("beta", []styledRuneRange{{start: 2, end: 3, prefix: "prefix-b"}})

	_, actualOK := subject.Get("alpha")
	if !actualOK {
		t.Fatal("expected alpha to be cached before eviction")
	}

	subject.Put("gamma", []styledRuneRange{{start: 3, end: 4, prefix: "prefix-c"}})

	if _, actualOK := subject.Get("beta"); actualOK {
		t.Fatal("expected beta to be evicted as the least recently used value")
	}

	actual, actualOK := subject.Get("alpha")
	if !actualOK {
		t.Fatal("expected alpha to stay cached after being read")
	}
	expected := []styledRuneRange{{start: 1, end: 2, prefix: "prefix-a"}}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected cached ranges %v, actual %v", expected, actual)
	}
}
