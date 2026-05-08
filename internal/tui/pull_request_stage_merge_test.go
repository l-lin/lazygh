package tui

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	persistcache "codeberg.org/l-lin/lazygh/internal/cache"
	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

func TestActionsPopup_GivenADraftPullRequestSelection_WhenOpening_ThenItShowsMarkReadyForReviewAndHidesConvertToDraftAndSquashMerge(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestLifecycleModel(given_pullRequestLifecycleSummary("OPEN", true)))
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if !strings.Contains(popupView.Buffer(), markPullRequestReadyForReviewActionTitle) {
		t.Fatalf("expected the popup to contain %q, actual %q", markPullRequestReadyForReviewActionTitle, popupView.Buffer())
	}
	for _, unexpected := range []string{convertPullRequestToDraftActionTitle, squashMergePullRequestActionTitle} {
		if strings.Contains(popupView.Buffer(), unexpected) {
			t.Fatalf("expected the popup to hide %q, actual %q", unexpected, popupView.Buffer())
		}
	}
}

func TestActionsPopup_GivenAnOpenPullRequestDescriptionDetail_WhenOpening_ThenItShowsConvertToDraftAndSquashMerge(t *testing.T) {
	model := given_pullRequestLifecycleModel(given_pullRequestLifecycleSummary("OPEN", false))
	model.OpenDetail()
	subject := NewProgramWithModel(model)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	for _, expected := range []string{convertPullRequestToDraftActionTitle, squashMergePullRequestActionTitle} {
		if !strings.Contains(popupView.Buffer(), expected) {
			t.Fatalf("expected the popup to contain %q, actual %q", expected, popupView.Buffer())
		}
	}
	if strings.Contains(popupView.Buffer(), markPullRequestReadyForReviewActionTitle) {
		t.Fatalf("expected the popup to hide %q, actual %q", markPullRequestReadyForReviewActionTitle, popupView.Buffer())
	}
}

func TestActionsPopup_GivenAnOpenPullRequestCommentsTab_WhenOpening_ThenItHidesTheStageToggleAndSquashMergeActions(t *testing.T) {
	model := given_pullRequestLifecycleModel(given_pullRequestLifecycleSummary("OPEN", false))
	model.OpenDetail()
	subject := NewProgramWithModel(model)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	for _, unexpected := range []string{markPullRequestReadyForReviewActionTitle, convertPullRequestToDraftActionTitle, squashMergePullRequestActionTitle} {
		if strings.Contains(popupView.Buffer(), unexpected) {
			t.Fatalf("expected the popup to hide %q away from the description tab, actual %q", unexpected, popupView.Buffer())
		}
	}
}

func TestLayout_GivenAReadyForReviewMutation_WhenRendering_ThenTheUpdatedOpenStateFeedbackAndCacheInvalidationAreVisible(t *testing.T) {
	summary := given_pullRequestLifecycleSummary("OPEN", true)
	model := given_pullRequestLifecycleModel(summary)
	model.OpenDetail()
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": given_pullRequestLifecycleDetail("OPEN", true),
		},
	}
	subject := given_pullRequestCommentProgram(model, loader)
	asyncRunner := &capturingAsyncRunner{}
	subject.asyncRunner = asyncRunner
	subject.uiUpdater = immediateUIUpdater{}
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: given_pullRequestLifecycleDetail("OPEN", true)}
	subject.pullRequestDiffCache["acme/widgets#42"] = pullRequestDiffResult{data: buildReviewDiffData(given_reviewSessionPullRequestDiff())}
	subject.reviewDiffRenderCache[reviewDiffRenderCacheKey{identity: "acme/widgets#42:main.go", width: 80}] = reviewDiffRenderCacheEntry{}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("mark ready", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "mark ready"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)

	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	if !reflect.DeepEqual(loader.markReadyForReviewCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected ready-for-review calls %v, actual %v", []string{"acme/widgets#42"}, loader.markReadyForReviewCalls)
	}
	then_currentViewNameIs(t, gui, viewDetailName)
	then_statusLineContains(t, gui, pullRequestMarkedReadyForReviewSuccessMessage)
	pullRequestsView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	if !strings.Contains(pullRequestsView.Buffer(), " acme/widgets#42 Lifecycle PR") {
		t.Fatalf("expected the pull-requests buffer to contain the open row, actual %q", pullRequestsView.Buffer())
	}
	if strings.Contains(pullRequestsView.Buffer(), " acme/widgets#42 Lifecycle PR") {
		t.Fatalf("expected the pull-requests buffer to drop the draft row, actual %q", pullRequestsView.Buffer())
	}
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "OPEN") {
		t.Fatalf("expected the detail buffer to contain %q, actual %q", "OPEN", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), "DRAFT") {
		t.Fatalf("expected the detail buffer to drop %q, actual %q", "DRAFT", detailView.Buffer())
	}
	if _, ok := subject.pullRequestDiffCache["acme/widgets#42"]; ok {
		t.Fatal("expected the cached pull-request diff to be invalidated after the stage mutation")
	}
	if len(subject.reviewDiffRenderCache) != 0 {
		t.Fatalf("expected the review diff render cache to be cleared, actual %d entries", len(subject.reviewDiffRenderCache))
	}
	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued detail refresh, actual %d", len(asyncRunner.runs))
	}
}

func TestActionsPopup_GivenASquashMergeAction_WhenExecutingOnce_ThenItAsksForConfirmationBeforeCallingGitHub(t *testing.T) {
	loader := &fakePullRequestDetailLoader{}
	subject := given_pullRequestCommentProgram(given_pullRequestLifecycleModel(given_pullRequestLifecycleSummary("OPEN", false)), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("squash", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "squash"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)

	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	then_currentViewNameIs(t, gui, viewActionsPopupName)
	if len(loader.squashMergeCalls) != 0 {
		t.Fatalf("expected no squash-merge calls before confirmation, actual %v", loader.squashMergeCalls)
	}
	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if !strings.Contains(popupView.Title, squashMergePullRequestConfirmationPromptMessage) {
		t.Fatalf("expected the popup title to contain %q, actual %q", squashMergePullRequestConfirmationPromptMessage, popupView.Title)
	}
}

func TestLayout_GivenAConfirmedSquashMerge_WhenRendering_ThenTheMergedStateFeedbackAndCacheInvalidationAreVisible(t *testing.T) {
	summary := given_pullRequestLifecycleSummary("OPEN", false)
	model := given_pullRequestLifecycleModel(summary)
	model.OpenDetail()
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": given_pullRequestLifecycleDetail("OPEN", false),
		},
	}
	subject := given_pullRequestCommentProgram(model, loader)
	asyncRunner := &capturingAsyncRunner{}
	subject.asyncRunner = asyncRunner
	subject.uiUpdater = immediateUIUpdater{}
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: given_pullRequestLifecycleDetail("OPEN", false)}
	subject.pullRequestDiffCache["acme/widgets#42"] = pullRequestDiffResult{data: buildReviewDiffData(given_reviewSessionPullRequestDiff())}
	subject.reviewDiffRenderCache[reviewDiffRenderCacheKey{identity: "acme/widgets#42:main.go", width: 80}] = reviewDiffRenderCacheEntry{}
	cache := &fakePersistentPullRequestCache{
		pullRequestsBySearchKey: map[string][]githubcli.PullRequest{},
		details:                 map[string]persistcache.CachedPullRequestDetail{},
		diffs:                   map[string]persistcache.CachedPullRequestDiff{},
	}
	subject.pullRequestCache = cache
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("squash", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "squash"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	if !reflect.DeepEqual(loader.squashMergeCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected squash-merge calls %v, actual %v", []string{"acme/widgets#42"}, loader.squashMergeCalls)
	}
	if !reflect.DeepEqual(cache.invalidatedPullRequests, []string{"acme/widgets#42"}) {
		t.Fatalf("expected invalidated pull requests %v, actual %v", []string{"acme/widgets#42"}, cache.invalidatedPullRequests)
	}
	then_currentViewNameIs(t, gui, viewDetailName)
	then_statusLineContains(t, gui, pullRequestSquashMergedSuccessMessage)
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "MERGED") {
		t.Fatalf("expected the detail buffer to contain %q, actual %q", "MERGED", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), "DRAFT") || strings.Contains(detailView.Buffer(), "OPEN") {
		t.Fatalf("expected the detail buffer to drop the pre-merge status, actual %q", detailView.Buffer())
	}
	if _, ok := subject.pullRequestDiffCache["acme/widgets#42"]; ok {
		t.Fatal("expected the cached pull-request diff to be invalidated after the squash merge")
	}
	if len(subject.reviewDiffRenderCache) != 0 {
		t.Fatalf("expected the review diff render cache to be cleared, actual %d entries", len(subject.reviewDiffRenderCache))
	}
	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued detail refresh, actual %d", len(asyncRunner.runs))
	}
}

func TestActionsPopup_GivenAConfirmedSquashMergeFailure_WhenExecuting_ThenItKeepsTheUIStableAndShowsTheGitHubError(t *testing.T) {
	loader := &fakePullRequestDetailLoader{squashMergeErr: errors.New("GitHub rejected the squash merge")}
	model := given_pullRequestLifecycleModel(given_pullRequestLifecycleSummary("OPEN", false))
	model.OpenDetail()
	subject := given_pullRequestCommentProgram(model, loader)
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: given_pullRequestLifecycleDetail("OPEN", false)}
	subject.pullRequestDiffCache["acme/widgets#42"] = pullRequestDiffResult{data: buildReviewDiffData(given_reviewSessionPullRequestDiff())}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("squash", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "squash"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	then_currentViewNameIs(t, gui, viewActionsPopupName)
	if !reflect.DeepEqual(loader.squashMergeCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected squash-merge calls %v, actual %v", []string{"acme/widgets#42"}, loader.squashMergeCalls)
	}
	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if !strings.Contains(popupView.Title, "GitHub rejected the squash merge") {
		t.Fatalf("expected the popup title to contain %q, actual %q", "GitHub rejected the squash merge", popupView.Title)
	}
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "OPEN") {
		t.Fatalf("expected the detail buffer to keep %q after the failure, actual %q", "OPEN", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), "MERGED") {
		t.Fatalf("expected the detail buffer to stay unmerged after the failure, actual %q", detailView.Buffer())
	}
	if _, ok := subject.pullRequestDiffCache["acme/widgets#42"]; !ok {
		t.Fatal("expected the cached pull-request diff to stay intact after the failed squash merge")
	}
}

func given_pullRequestLifecycleModel(summary githubcli.PullRequest) *Model {
	model := given_model()
	model.FocusPullRequestsView()
	model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{myPullRequestRow(summary)})
	return model
}

func given_pullRequestLifecycleSummary(state string, isDraft bool) githubcli.PullRequest {
	return githubcli.PullRequest{
		Title:      "Lifecycle PR",
		Number:     42,
		Repository: githubcli.Repository{NameWithOwner: "acme/widgets"},
		URL:        "https://github.com/acme/widgets/pull/42",
		Body:       "Lifecycle body",
		State:      state,
		IsDraft:    isDraft,
	}
}

func given_pullRequestLifecycleDetail(state string, isDraft bool) githubcli.PullRequestDetail {
	return githubcli.PullRequestDetail{
		Title:       "Lifecycle PR",
		Number:      42,
		Body:        "Lifecycle body",
		BaseRefName: "main",
		HeadRefName: "feature/lifecycle",
		State:       state,
		IsDraft:     isDraft,
	}
}
