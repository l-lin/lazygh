package githubcli

import (
	"reflect"
	"testing"
)

func TestNormalizeUniquePullRequestAuthors_GivenWhitespaceAndDuplicates_WhenNormalizing_ThenItKeepsTheFirstSeenOrder(t *testing.T) {
	actual := normalizeUniquePullRequestAuthors([]PullRequestAuthor{
		{Login: " bob ", Name: " Bob ", IsBot: false},
		{Login: "alice", Name: " Alice ", IsBot: false},
		{Login: "bob", Name: "Duplicate", IsBot: true},
	})

	expected := []PullRequestAuthor{
		{Login: "bob", Name: "Bob", IsBot: false},
		{Login: "alice", Name: "Alice", IsBot: false},
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected normalized authors %+v, actual %+v", expected, actual)
	}
}

func TestNormalizeUniqueStrings_GivenWhitespaceAndDuplicates_WhenNormalizing_ThenItKeepsTheFirstSeenValues(t *testing.T) {
	actual := normalizeUniqueStrings([]string{" alice ", "bob", "alice", "", " bob "})

	expected := []string{"alice", "bob"}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected normalized strings %v, actual %v", expected, actual)
	}
}
