package githubcli

import (
	"errors"
	"reflect"
	"testing"
)

func TestParseRepositoryMergeCapabilities_GivenGraphQLResponse_WhenParsing_ThenItReturnsTheNormalizedCapabilities(t *testing.T) {
	subject := []byte(`{"data":{"repository":{"autoMergeAllowed":true}}}`)

	actual, actualErr := parseRepositoryMergeCapabilities(subject)

	then_noError(t, actualErr)
	expected := repositoryMergeCapabilities{AutoMergeAllowed: true}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected repository capabilities %+v, actual %+v", expected, actual)
	}
}

func TestLoadRepositoryMergeCapabilities_GivenRepository_WhenFetching_ThenItRunsTheGraphQLQuery(t *testing.T) {
	runner := &fakeRunner{stdout: []byte(`{"data":{"repository":{"autoMergeAllowed":true}}}`)}
	subject := NewPullRequestMutationServiceWithRunner(runner)

	actual, actualErr := subject.loadRepositoryMergeCapabilities("acme/widgets")

	then_noError(t, actualErr)
	expected := repositoryMergeCapabilities{AutoMergeAllowed: true}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected repository capabilities %+v, actual %+v", expected, actual)
	}
	then_commandIs(t, runner, "gh", []string{"api", "graphql", "-f", "query=" + repositoryMergeCapabilitiesQuery, "-F", "owner=acme", "-F", "name=widgets"})
}

func TestParseRepositoryMergeCapabilities_GivenMissingRepository_WhenParsing_ThenItReturnsTheInvalidResponseError(t *testing.T) {
	_, actualErr := parseRepositoryMergeCapabilities([]byte(`{"data":{"repository":null}}`))

	if !errors.Is(actualErr, ErrInvalidRepositoryMergeCapabilitiesResponse) {
		t.Fatalf("expected error %v, actual %v", ErrInvalidRepositoryMergeCapabilitiesResponse, actualErr)
	}
}
