package githubcli

import (
	"errors"
	"reflect"
	"testing"
)

func TestGetPullRequestSummary_GivenRemoteMetadata_WhenProbing_ThenItUsesOneViewRequest(t *testing.T) {
	runner := &fakeRunner{stdout: []byte(`{"number":42,"title":"Changed","updatedAt":"2026-10-04T12:00:00Z","state":"OPEN"}`)}
	subject := NewPullRequestDetailServiceWithRunner(runner)
	actual, actualErr := subject.GetPullRequestSummary(" acme/widgets ", 42)
	expected := []string{"pr", "view", "42", "-R", "acme/widgets", "--json", "number,title,url,body,state,isDraft,updatedAt"}
	if actualErr != nil || actual.UpdatedAt != "2026-10-04T12:00:00Z" || actual.Repository.NameWithOwner != "acme/widgets" || !reflect.DeepEqual(runner.args, expected) {
		t.Fatalf("expected remote summary and command %v, actual %+v args=%v err=%v", expected, actual, runner.args, actualErr)
	}
}

func TestGetPullRequestSummary_GivenMissingVersion_WhenProbing_ThenItRejectsTheResponse(t *testing.T) {
	subject := NewPullRequestDetailServiceWithRunner(&fakeRunner{stdout: []byte(`{"number":42}`)})
	_, actual := subject.GetPullRequestSummary("acme/widgets", 42)
	expected := ErrInvalidPullRequestDetailResponse
	if !errors.Is(actual, expected) {
		t.Fatalf("expected %v, actual %v", expected, actual)
	}
}

func TestGetPullRequestSummary_GivenInvalidJSON_WhenProbing_ThenItReturnsAnError(t *testing.T) {
	subject := NewPullRequestDetailServiceWithRunner(&fakeRunner{stdout: []byte(`{"number":`)})
	_, actual := subject.GetPullRequestSummary("acme/widgets", 42)
	expected := ErrInvalidPullRequestDetailResponse
	if !errors.Is(actual, expected) {
		t.Fatalf("expected %v, actual %v", expected, actual)
	}
}
