package github

import "testing"

func TestRepositoryShortName_GivenRepositoryName_WhenDeriving_ThenItPrefersTheTrimmedName(t *testing.T) {
	subject := RepositoryRef{Name: " widgets ", NameWithOwner: "acme/ignored"}

	actual := RepositoryShortName(subject)
	expected := "widgets"
	if actual != expected {
		t.Fatalf("expected short repository name %q, actual %q", expected, actual)
	}
}

func TestRepositoryShortName_GivenMissingName_WhenDeriving_ThenItFallsBackToTheLastOwnerSegment(t *testing.T) {
	subject := RepositoryRef{NameWithOwner: " acme/platform/widgets "}

	actual := RepositoryShortName(subject)
	expected := "widgets"
	if actual != expected {
		t.Fatalf("expected short repository name %q, actual %q", expected, actual)
	}
}

func TestRepositoryShortName_GivenMissingUsableFields_WhenDeriving_ThenItReturnsEmpty(t *testing.T) {
	testCases := []struct {
		name       string
		repository RepositoryRef
	}{
		{name: "blank fields", repository: RepositoryRef{}},
		{name: "blank owner segment", repository: RepositoryRef{NameWithOwner: " / / "}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actual := RepositoryShortName(testCase.repository)
			expected := ""
			if actual != expected {
				t.Fatalf("expected short repository name %q, actual %q", expected, actual)
			}
		})
	}
}
