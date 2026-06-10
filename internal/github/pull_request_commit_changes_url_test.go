package github

import "testing"

func TestPullRequestCommitChangesURL_GivenValidPullRequestIdentity_WhenBuilding_ThenItReturnsTheCommitChangesURL(t *testing.T) {
	repository := Repository{NameWithOwner: "acme/widgets"}

	actual, ok := PullRequestCommitChangesURL(repository, 42, "2222222bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")

	if !ok {
		t.Fatal("expected a commit changes URL")
	}
	expected := "https://github.com/acme/widgets/pull/42/changes/2222222bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	if actual != expected {
		t.Fatalf("expected commit changes URL %q, actual %q", expected, actual)
	}
}

func TestPullRequestCommitChangesURL_GivenMissingRepositoryNumberOrCommitOID_WhenBuilding_ThenItReturnsNoURL(t *testing.T) {
	testCases := []struct {
		name       string
		repository Repository
		number     int
		commitOID  string
	}{
		{name: "missing repository", repository: Repository{}, number: 42, commitOID: "2222222bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"},
		{name: "missing number", repository: Repository{NameWithOwner: "acme/widgets"}, number: 0, commitOID: "2222222bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"},
		{name: "missing commit oid", repository: Repository{NameWithOwner: "acme/widgets"}, number: 42, commitOID: "  "},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actual, ok := PullRequestCommitChangesURL(testCase.repository, testCase.number, testCase.commitOID)

			if ok {
				t.Fatalf("expected no URL, actual %q", actual)
			}
			if actual != "" {
				t.Fatalf("expected an empty URL, actual %q", actual)
			}
		})
	}
}
