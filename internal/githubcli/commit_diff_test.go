package githubcli

import (
	"errors"
	"reflect"
	"testing"
)

func TestGetCommitDiff_GivenAValidCommitResponse_WhenFetching_ThenItReturnsNormalizedFiles(t *testing.T) {
	runner := &fakeRunner{responses: []fakeCommandResponse{{stdout: []byte(`{"files":[{"filename":" internal/tui/render.go ","status":" modified ","additions":2,"deletions":1,"patch":"@@ -1 +1,2 @@\n-old\n+new\n+more"},{"filename":" docs/guide.md ","previous_filename":" docs/old-guide.md ","status":" renamed ","additions":0,"deletions":0}]}`)}}}
	subject := NewPullRequestDetailServiceWithRunner(runner)

	actual, actualErr := subject.GetCommitDiff(" acme/widgets ", " 1111111abcdef ")

	then_noError(t, actualErr)
	then_commandsAre(t, runner, []fakeCommandCall{{name: "gh", args: []string{"api", "repos/acme/widgets/commits/1111111abcdef"}}})

	expected := CommitDiff{Files: []PullRequestDiffFile{
		{Path: "internal/tui/render.go", ChangeType: "modified", Additions: 2, Deletions: 1, Patch: "@@ -1 +1,2 @@\n-old\n+new\n+more"},
		{Path: "docs/guide.md", PreviousPath: "docs/old-guide.md", ChangeType: "renamed"},
	}}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected commit diff %+v, actual %+v", expected, actual)
	}
}

func TestGetCommitDiff_GivenInvalidJSON_WhenFetching_ThenItReturnsAnInvalidResponseError(t *testing.T) {
	runner := &fakeRunner{responses: []fakeCommandResponse{{stdout: []byte(`{"files":`)}}}
	subject := NewPullRequestDetailServiceWithRunner(runner)

	_, actualErr := subject.GetCommitDiff("acme/widgets", "1111111abcdef")

	if !errors.Is(actualErr, ErrInvalidCommitDiffResponse) {
		t.Fatalf("expected error %v, actual %v", ErrInvalidCommitDiffResponse, actualErr)
	}
}
