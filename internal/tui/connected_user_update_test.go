package tui

import (
	"reflect"
	"testing"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func TestUpdate_GivenMsgConnectedUserLoadPlanned_WhenApplying_ThenItStartsTheSessionStoreLoadingState(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	subject.connectedUserLogin = "stale-login"
	subject.connectedUserName = "stale name"

	Update(subject, MsgConnectedUserLoadPlanned{})

	if !subject.connectedUserLoadStarted {
		t.Fatal("expected connected-user workflow planning to mark the load as started")
	}
	if actual := subject.connectedUserLogin; actual != "stale-login" {
		t.Fatalf("expected planned login %q, actual %q", "stale-login", actual)
	}
	if actual := subject.connectedUserName; actual != "stale name" {
		t.Fatalf("expected planned name %q, actual %q", "stale name", actual)
	}
}

func TestUpdate_GivenMsgConnectedUserLoadedWithADifferentLogin_WhenApplying_ThenItReplacesTheSessionStoreInvalidatesDependentCachesAndUpdatesTheUserRows(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	subject.connectedUserLoadStarted = true
	subject.connectedUserLogin = "reviewer"
	subject.connectedUserName = "Old Reviewer"
	given_connectedUserDependentCaches(subject)
	given_user := githubdomain.ConnectedUser{Login: " octocat ", Name: " Octo Cat "}

	Update(subject, MsgConnectedUserLoaded{User: given_user})

	if !subject.connectedUserLoadStarted {
		t.Fatal("expected the connected-user load-started flag to stay true after loading")
	}
	if actual := subject.connectedUserLogin; actual != "octocat" {
		t.Fatalf("expected loaded login %q, actual %q", "octocat", actual)
	}
	if actual := subject.connectedUserName; actual != "Octo Cat" {
		t.Fatalf("expected loaded name %q, actual %q", "Octo Cat", actual)
	}
	if actual := len(subject.pullRequestDetailDocumentCache); actual != 0 {
		t.Fatalf("expected the detail document cache to be invalidated, actual size %d", actual)
	}
	if actual := len(subject.pullRequestConversationDocumentCache); actual != 0 {
		t.Fatalf("expected the conversation document cache to be invalidated, actual size %d", actual)
	}
	if actual := len(subject.pullRequestChangesRenderedRowsCache); actual != 0 {
		t.Fatalf("expected the rendered changes cache to be invalidated, actual size %d", actual)
	}
	if actual := len(subject.reviewDiffRenderCache); actual != 0 {
		t.Fatalf("expected the review diff render cache to be invalidated, actual size %d", actual)
	}
	actual := subject.model.Users()
	expected := []Item{connectedUserStateItem(given_user, nil)}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected rendered users %+v, actual %+v", expected, actual)
	}
}

func TestUpdate_GivenMsgConnectedUserLoadedWithTheSameLogin_WhenApplying_ThenItKeepsDependentCachesAndStillUpdatesTheRenderedUserRow(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	subject.connectedUserLoadStarted = true
	subject.connectedUserLogin = "octocat"
	subject.connectedUserName = "Old Name"
	given_detailCacheKey, given_reviewCacheKey := given_connectedUserDependentCaches(subject)
	given_user := githubdomain.ConnectedUser{Login: "octocat", Name: "New Name"}

	Update(subject, MsgConnectedUserLoaded{User: given_user})

	if actual := subject.connectedUserLogin; actual != "octocat" {
		t.Fatalf("expected login %q, actual %q", "octocat", actual)
	}
	if actual := subject.connectedUserName; actual != "New Name" {
		t.Fatalf("expected name %q, actual %q", "New Name", actual)
	}
	if actual := len(subject.pullRequestDetailDocumentCache); actual != 1 {
		t.Fatalf("expected the detail document cache to stay intact, actual size %d", actual)
	}
	if _, ok := subject.pullRequestDetailDocumentCache[given_detailCacheKey]; !ok {
		t.Fatalf("expected the detail document cache key %+v to stay present", given_detailCacheKey)
	}
	if actual := len(subject.reviewDiffRenderCache); actual != 1 {
		t.Fatalf("expected the review diff render cache to stay intact, actual size %d", actual)
	}
	if _, ok := subject.reviewDiffRenderCache[given_reviewCacheKey]; !ok {
		t.Fatalf("expected the review diff render cache key %+v to stay present", given_reviewCacheKey)
	}
	actual := subject.model.Users()
	expected := []Item{connectedUserStateItem(given_user, nil)}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected rendered users %+v, actual %+v", expected, actual)
	}
}

func given_connectedUserDependentCaches(subject *Program) (pullRequestDetailDocumentCacheKey, reviewDiffRenderCacheKey) {
	detailCacheKey := pullRequestDetailDocumentCacheKey{pullRequestKey: "acme/widgets#42", tab: DescriptionDetailTab, width: 80}
	reviewCacheKey := reviewDiffRenderCacheKey{repositoryName: "acme/widgets", pullRequestNumber: 42, filePath: "README.md", width: 80}
	subject.pullRequestDetailDocumentCache[detailCacheKey] = detailDocument{}
	subject.pullRequestConversationDocumentCache[detailCacheKey] = browserConversationDocument{}
	subject.pullRequestChangesRenderedRowsCache[detailCacheKey] = []reviewDiffRenderedRow{{Text: "stale"}}
	subject.reviewDiffRenderCache[reviewCacheKey] = reviewDiffRenderCacheEntry{rows: []reviewDiffRenderedRow{{Text: "stale"}}}
	return detailCacheKey, reviewCacheKey
}
