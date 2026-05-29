package tui

import "testing"

func TestImageLoadCoordinator_GivenLoadFailureAndAuthTokenState_WhenApplyingTransitions_ThenItTracksStateThroughHelpers(t *testing.T) {
	subject := newImageLoadCoordinator(nil, nil)
	subject.detailImageHTMLLoadFailed["html:stale"] = true
	subject.detailImageLoadFailed["https://example.com/stale.png"] = true

	subject.markDetailImageHTMLLoadPlanned("html:success")
	subject.markDetailImageHTMLLoadPlanned("html:failed")
	subject.markDetailImageLoadPlanned("https://example.com/diagram.png")
	subject.markDetailImageLoadPlanned("https://example.com/fail.png")
	subject.recordDetailImageHTMLLoadFinished("html:success", false)
	subject.recordDetailImageHTMLLoadFinished("html:failed", true)
	subject.recordDetailImageLoadFinished("https://example.com/diagram.png", false)
	subject.recordDetailImageLoadFinished("https://example.com/fail.png", true)
	subject.cacheGitHubAuthToken(" ghp_secret-token ")

	actualToken, actualLoaded := subject.cachedGitHubAuthToken()
	if !actualLoaded {
		t.Fatal("expected the coordinator to remember that the auth token was loaded")
	}
	if actualToken != "ghp_secret-token" {
		t.Fatalf("expected cached auth token %q, actual %q", "ghp_secret-token", actualToken)
	}
	if subject.detailImageHTMLLoadInFlight["html:success"] {
		t.Fatal("expected successful HTML loads to clear the in-flight state")
	}
	if !subject.detailImageHTMLLoadFailed["html:failed"] {
		t.Fatal("expected failed HTML loads to be tracked")
	}
	if subject.detailImageHTMLLoadFailed["html:success"] {
		t.Fatal("expected successful HTML loads to clear stale failure state")
	}
	if subject.detailImageLoadInFlight["https://example.com/diagram.png"] {
		t.Fatal("expected successful image loads to clear the in-flight state")
	}
	if !subject.detailImageLoadFailed["https://example.com/fail.png"] {
		t.Fatal("expected failed image loads to be tracked")
	}
	if subject.detailImageLoadFailed["https://example.com/diagram.png"] {
		t.Fatal("expected successful image loads to clear stale failure state")
	}
}
