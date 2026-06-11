package tui

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	persistcache "github.com/l-lin/lazygh/internal/cache"
	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestActionsPopup_GivenADraftPullRequestSelection_WhenOpening_ThenItShowsMarkReadyForReviewAndClosePRAndHidesConvertToDraftAndSquashMerge(t *testing.T) {
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
	for _, expected := range []string{markPullRequestReadyForReviewActionTitle, closePullRequestActionTitle} {
		if !strings.Contains(popupView.Buffer(), expected) {
			t.Fatalf("expected the popup to contain %q, actual %q", expected, popupView.Buffer())
		}
	}
	for _, unexpected := range []string{convertPullRequestToDraftActionTitle, squashMergePullRequestActionTitle} {
		if strings.Contains(popupView.Buffer(), unexpected) {
			t.Fatalf("expected the popup to hide %q, actual %q", unexpected, popupView.Buffer())
		}
	}
}

func TestActionsPopup_GivenAnOpenPullRequestDescriptionDetail_WhenOpening_ThenItShowsConvertToDraftSquashMergeAndClosePR(t *testing.T) {
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
	for _, expected := range []string{convertPullRequestToDraftActionTitle, squashMergePullRequestActionTitle, closePullRequestActionTitle} {
		if !strings.Contains(popupView.Buffer(), expected) {
			t.Fatalf("expected the popup to contain %q, actual %q", expected, popupView.Buffer())
		}
	}
	if strings.Contains(popupView.Buffer(), markPullRequestReadyForReviewActionTitle) {
		t.Fatalf("expected the popup to hide %q, actual %q", markPullRequestReadyForReviewActionTitle, popupView.Buffer())
	}
}

func TestActionsPopup_GivenAnOpenPullRequestDescriptionDetailWithAnOngoingBuild_WhenOpening_ThenItShowsEnableAutoMergeAndHidesSquashMerge(t *testing.T) {
	summary := given_pullRequestLifecycleSummary("OPEN", false)
	summary.StatusCheckRollupState = "PENDING"
	model := given_pullRequestLifecycleModel(summary)
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
	if !strings.Contains(popupView.Buffer(), enablePullRequestAutoMergeActionTitle) {
		t.Fatalf("expected the popup to contain %q, actual %q", enablePullRequestAutoMergeActionTitle, popupView.Buffer())
	}
	for _, unexpected := range []string{disablePullRequestAutoMergeActionTitle, squashMergePullRequestActionTitle} {
		if strings.Contains(popupView.Buffer(), unexpected) {
			t.Fatalf("expected the popup to hide %q, actual %q", unexpected, popupView.Buffer())
		}
	}
}

func TestActionsPopup_GivenAnOpenPullRequestDescriptionDetailWithAnOngoingBuildAndAutoMergeEnabled_WhenOpening_ThenItShowsDisableAutoMergeAndHidesSquashMerge(t *testing.T) {
	summary := given_pullRequestLifecycleSummary("OPEN", false)
	summary.StatusCheckRollupState = "PENDING"
	summary.AutoMergeRequest = &githubcli.PullRequestAutoMergeRequest{EnabledAt: "2026-05-20T10:00:00Z"}
	model := given_pullRequestLifecycleModel(summary)
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
	if !strings.Contains(popupView.Buffer(), disablePullRequestAutoMergeActionTitle) {
		t.Fatalf("expected the popup to contain %q, actual %q", disablePullRequestAutoMergeActionTitle, popupView.Buffer())
	}
	for _, unexpected := range []string{enablePullRequestAutoMergeActionTitle, squashMergePullRequestActionTitle} {
		if strings.Contains(popupView.Buffer(), unexpected) {
			t.Fatalf("expected the popup to hide %q, actual %q", unexpected, popupView.Buffer())
		}
	}
}

func TestActionsPopup_GivenAnOpenPullRequestDescriptionDetailWaitingForRequiredReview_WhenOpening_ThenItShowsEnableAutoMergeAndHidesSquashMerge(t *testing.T) {
	summary := given_pullRequestLifecycleSummary("OPEN", false)
	summary.ReviewDecision = "REVIEW_REQUIRED"
	summary.Mergeable = "MERGEABLE"
	summary.MergeStateStatus = "CLEAN"
	model := given_pullRequestLifecycleModel(summary)
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
	if !strings.Contains(popupView.Buffer(), enablePullRequestAutoMergeActionTitle) {
		t.Fatalf("expected the popup to contain %q, actual %q", enablePullRequestAutoMergeActionTitle, popupView.Buffer())
	}
	for _, unexpected := range []string{disablePullRequestAutoMergeActionTitle, squashMergePullRequestActionTitle} {
		if strings.Contains(popupView.Buffer(), unexpected) {
			t.Fatalf("expected the popup to hide %q, actual %q", unexpected, popupView.Buffer())
		}
	}
}

func TestActionsPopup_GivenAnOpenPullRequestDescriptionDetailWaitingForRequiredReviewWithAutoMergeEnabled_WhenOpening_ThenItShowsDisableAutoMergeAndHidesSquashMerge(t *testing.T) {
	summary := given_pullRequestLifecycleSummary("OPEN", false)
	summary.ReviewDecision = "REVIEW_REQUIRED"
	summary.Mergeable = "MERGEABLE"
	summary.MergeStateStatus = "CLEAN"
	summary.AutoMergeRequest = &githubcli.PullRequestAutoMergeRequest{EnabledAt: "2026-05-20T10:00:00Z"}
	model := given_pullRequestLifecycleModel(summary)
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
	if !strings.Contains(popupView.Buffer(), disablePullRequestAutoMergeActionTitle) {
		t.Fatalf("expected the popup to contain %q, actual %q", disablePullRequestAutoMergeActionTitle, popupView.Buffer())
	}
	for _, unexpected := range []string{enablePullRequestAutoMergeActionTitle, squashMergePullRequestActionTitle} {
		if strings.Contains(popupView.Buffer(), unexpected) {
			t.Fatalf("expected the popup to hide %q, actual %q", unexpected, popupView.Buffer())
		}
	}
}

func TestActionsPopup_GivenALoadedDetailWithoutAutoMerge_WhenOpening_ThenItUsesTheLoadedStateInsteadOfTheStaleSummary(t *testing.T) {
	summary := given_pullRequestLifecycleSummary("OPEN", false)
	summary.StatusCheckRollupState = "PENDING"
	summary.AutoMergeRequest = given_pullRequestAutoMergeRequest()
	model := given_pullRequestLifecycleModel(summary)
	model.OpenDetail()
	subject := NewProgramWithModel(model)
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(given_pullRequestLifecycleDetailWithPendingBuild("OPEN", false, false))}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if !strings.Contains(popupView.Buffer(), enablePullRequestAutoMergeActionTitle) {
		t.Fatalf("expected the popup to contain %q, actual %q", enablePullRequestAutoMergeActionTitle, popupView.Buffer())
	}
	for _, unexpected := range []string{disablePullRequestAutoMergeActionTitle, squashMergePullRequestActionTitle} {
		if strings.Contains(popupView.Buffer(), unexpected) {
			t.Fatalf("expected the popup to hide %q, actual %q", unexpected, popupView.Buffer())
		}
	}
}

func TestActionsPopup_GivenAnOpenPullRequestDescriptionDetailWithACompletedBuildAndAutoMergeEnabled_WhenOpening_ThenItShowsSquashMergeAndHidesTheAutoMergeToggle(t *testing.T) {
	summary := given_pullRequestLifecycleSummary("OPEN", false)
	summary.StatusCheckRollupState = "SUCCESS"
	summary.AutoMergeRequest = &githubcli.PullRequestAutoMergeRequest{EnabledAt: "2026-05-20T10:00:00Z"}
	model := given_pullRequestLifecycleModel(summary)
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
	if !strings.Contains(popupView.Buffer(), squashMergePullRequestActionTitle) {
		t.Fatalf("expected the popup to contain %q, actual %q", squashMergePullRequestActionTitle, popupView.Buffer())
	}
	for _, unexpected := range []string{enablePullRequestAutoMergeActionTitle, disablePullRequestAutoMergeActionTitle} {
		if strings.Contains(popupView.Buffer(), unexpected) {
			t.Fatalf("expected the popup to hide %q, actual %q", unexpected, popupView.Buffer())
		}
	}
}

func TestActionsPopup_GivenAnOutOfDateOpenPullRequestDescriptionDetail_WhenOpening_ThenItShowsUpdateBranch(t *testing.T) {
	model := given_pullRequestLifecycleModel(given_pullRequestLifecycleSummary("OPEN", false))
	model.OpenDetail()
	subject := NewProgramWithModel(model)
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(given_pullRequestOutOfDateLifecycleDetail("OPEN", false))}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if !strings.Contains(popupView.Buffer(), updatePullRequestBranchActionTitle) {
		t.Fatalf("expected the popup to contain %q, actual %q", updatePullRequestBranchActionTitle, popupView.Buffer())
	}
}

func TestActionsPopup_GivenAnUpToDateOpenPullRequestDescriptionDetail_WhenOpening_ThenItHidesUpdateBranch(t *testing.T) {
	model := given_pullRequestLifecycleModel(given_pullRequestLifecycleSummary("OPEN", false))
	model.OpenDetail()
	subject := NewProgramWithModel(model)
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(given_pullRequestLifecycleDetail("OPEN", false))}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if strings.Contains(popupView.Buffer(), updatePullRequestBranchActionTitle) {
		t.Fatalf("expected the popup to hide %q, actual %q", updatePullRequestBranchActionTitle, popupView.Buffer())
	}
}

func TestActionsPopup_GivenAnOpenPullRequestCommentsTab_WhenOpening_ThenItHidesTheLifecycleActions(t *testing.T) {
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
	for _, unexpected := range []string{markPullRequestReadyForReviewActionTitle, convertPullRequestToDraftActionTitle, squashMergePullRequestActionTitle, closePullRequestActionTitle} {
		if strings.Contains(popupView.Buffer(), unexpected) {
			t.Fatalf("expected the popup to hide %q away from the description tab, actual %q", unexpected, popupView.Buffer())
		}
	}
}

func TestActionsPopup_GivenAClosedDraftPullRequestDescriptionDetail_WhenOpening_ThenItShowsReopenPRAndHidesOtherLifecycleActions(t *testing.T) {
	model := given_pullRequestLifecycleModel(given_pullRequestLifecycleSummary("CLOSED", true))
	model.OpenDetail()
	subject := NewProgramWithModel(model)
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(given_pullRequestLifecycleDetail("CLOSED", true))}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if !strings.Contains(popupView.Buffer(), reopenPullRequestActionTitle) {
		t.Fatalf("expected the popup to contain %q, actual %q", reopenPullRequestActionTitle, popupView.Buffer())
	}
	for _, unexpected := range []string{markPullRequestReadyForReviewActionTitle, convertPullRequestToDraftActionTitle, squashMergePullRequestActionTitle, closePullRequestActionTitle} {
		if strings.Contains(popupView.Buffer(), unexpected) {
			t.Fatalf("expected the popup to hide %q for a closed draft pull request, actual %q", unexpected, popupView.Buffer())
		}
	}
}

func TestActionsPopup_GivenAReadyForReviewMutation_WhenExecuting_ThenItKeepsThePopupOpenShowsTheStatusLineSpinnerAndDelaysTheGitHubCall(t *testing.T) {
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
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(given_pullRequestLifecycleDetail("OPEN", true))}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch(markPullRequestReadyForReviewActionTitle, matchingActionsPopupIndexes(subject.currentActionsPopupActions(), markPullRequestReadyForReviewActionTitle))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)

	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued ready-for-review mutation, actual %d", len(asyncRunner.runs))
	}
	if len(loader.markReadyForReviewCalls) != 0 {
		t.Fatalf("expected the ready-for-review call to wait for the queued run, actual %v", loader.markReadyForReviewCalls)
	}
	then_currentViewNameIs(t, gui, viewActionsPopupSearchName)
	_, actualErr = gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	then_statusLineContains(t, gui, string(loadingSpinnerFrames[0]))
	then_statusLineContains(t, gui, "Running `gh pr ready 42 -R acme/widgets`.")
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "DRAFT") {
		t.Fatalf("expected the detail buffer to keep %q while the mutation is in flight, actual %q", "DRAFT", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), "PR marked ready for review") {
		t.Fatalf("expected the detail buffer to avoid the success message while the mutation is in flight, actual %q", detailView.Buffer())
	}
}

func TestActionsPopup_GivenAConvertToDraftMutation_WhenExecuting_ThenItKeepsThePopupOpenShowsTheStatusLineSpinnerAndDelaysTheGitHubCall(t *testing.T) {
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
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(given_pullRequestLifecycleDetail("OPEN", false))}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch(convertPullRequestToDraftActionTitle, matchingActionsPopupIndexes(subject.currentActionsPopupActions(), convertPullRequestToDraftActionTitle))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)

	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued convert-to-draft mutation, actual %d", len(asyncRunner.runs))
	}
	if len(loader.convertToDraftCalls) != 0 {
		t.Fatalf("expected the convert-to-draft call to wait for the queued run, actual %v", loader.convertToDraftCalls)
	}
	then_currentViewNameIs(t, gui, viewActionsPopupSearchName)
	_, actualErr = gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	then_statusLineContains(t, gui, string(loadingSpinnerFrames[0]))
	then_statusLineContains(t, gui, "Running `gh pr ready 42 -R acme/widgets --undo`.")
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "OPEN") || strings.Contains(detailView.Buffer(), "DRAFT") {
		t.Fatalf("expected the detail buffer to keep the open non-draft state while the mutation is in flight, actual %q", detailView.Buffer())
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
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(given_pullRequestLifecycleDetail("OPEN", true))}
	subject.pullRequestDiffCache["acme/widgets#42"] = pullRequestDiffResult{data: buildReviewDiffData(given_reviewSessionPullRequestDiff())}
	subject.reviewDiffRenderCache[reviewDiffRenderCacheKey{repositoryName: "acme/widgets", pullRequestNumber: 42, filePath: "main.go", width: 80}] = reviewDiffRenderCacheEntry{}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch(markPullRequestReadyForReviewActionTitle, matchingActionsPopupIndexes(subject.currentActionsPopupActions(), markPullRequestReadyForReviewActionTitle))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)

	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued ready-for-review mutation, actual %d", len(asyncRunner.runs))
	}
	asyncRunner.runs[0]()

	if !reflect.DeepEqual(loader.markReadyForReviewCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected ready-for-review calls %v, actual %v", []string{"acme/widgets#42"}, loader.markReadyForReviewCalls)
	}
	then_currentViewNameIs(t, gui, viewDetailName)
	then_statusLineContains(t, gui, pullRequestMarkedReadyForReviewSuccessMessage)
	pullRequestsView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	if !strings.Contains(pullRequestsView.Buffer(), " widgets#42 Lifecycle PR") {
		t.Fatalf("expected the pull-requests buffer to contain the open row, actual %q", pullRequestsView.Buffer())
	}
	if strings.Contains(pullRequestsView.Buffer(), " widgets#42 Lifecycle PR") {
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
	if len(asyncRunner.runs) != 2 {
		t.Fatalf("expected one queued ready-for-review mutation plus one queued detail refresh, actual %d", len(asyncRunner.runs))
	}
}

func TestLayout_GivenAClosePullRequestMutation_WhenRendering_ThenTheUpdatedClosedStateFeedbackAndCacheInvalidationAreVisible(t *testing.T) {
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
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(given_pullRequestLifecycleDetail("OPEN", false))}
	subject.pullRequestDiffCache["acme/widgets#42"] = pullRequestDiffResult{data: buildReviewDiffData(given_reviewSessionPullRequestDiff())}
	subject.reviewDiffRenderCache[reviewDiffRenderCacheKey{repositoryName: "acme/widgets", pullRequestNumber: 42, filePath: "main.go", width: 80}] = reviewDiffRenderCacheEntry{}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch(closePullRequestActionTitle, matchingActionsPopupIndexes(subject.currentActionsPopupActions(), closePullRequestActionTitle))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)

	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued close mutation, actual %d", len(asyncRunner.runs))
	}
	asyncRunner.runs[0]()

	if !reflect.DeepEqual(loader.closePullRequestCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected close pull request calls %v, actual %v", []string{"acme/widgets#42"}, loader.closePullRequestCalls)
	}
	then_currentViewNameIs(t, gui, viewDetailName)
	then_statusLineContains(t, gui, pullRequestClosedSuccessMessage)
	pullRequestsView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	if !strings.Contains(pullRequestsView.Buffer(), " widgets#42 Lifecycle PR") {
		t.Fatalf("expected the pull-requests buffer to keep the selected row, actual %q", pullRequestsView.Buffer())
	}
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "CLOSED") {
		t.Fatalf("expected the detail buffer to contain %q, actual %q", "CLOSED", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), "DRAFT") || strings.Contains(detailView.Buffer(), "OPEN") {
		t.Fatalf("expected the detail buffer to drop the open state, actual %q", detailView.Buffer())
	}
	if _, ok := subject.pullRequestDiffCache["acme/widgets#42"]; ok {
		t.Fatal("expected the cached pull-request diff to be invalidated after closing the pull request")
	}
	if len(subject.reviewDiffRenderCache) != 0 {
		t.Fatalf("expected the review diff render cache to be cleared, actual %d entries", len(subject.reviewDiffRenderCache))
	}
	if len(asyncRunner.runs) != 2 {
		t.Fatalf("expected one queued close mutation plus one queued detail refresh, actual %d", len(asyncRunner.runs))
	}
}

func TestLayout_GivenAnUpdateBranchMutation_WhenRendering_ThenItRemovesTheOutOfDateIndicatorAndQueuesARefresh(t *testing.T) {
	summary := given_pullRequestLifecycleSummary("OPEN", false)
	model := given_pullRequestLifecycleModel(summary)
	model.OpenDetail()
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": given_pullRequestOutOfDateLifecycleDetail("OPEN", false),
		},
	}
	subject := given_pullRequestCommentProgram(model, loader)
	asyncRunner := &capturingAsyncRunner{}
	subject.asyncRunner = asyncRunner
	subject.uiUpdater = immediateUIUpdater{}
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(given_pullRequestOutOfDateLifecycleDetail("OPEN", false))}
	subject.pullRequestDiffCache["acme/widgets#42"] = pullRequestDiffResult{data: buildReviewDiffData(given_reviewSessionPullRequestDiff())}
	subject.reviewDiffRenderCache[reviewDiffRenderCacheKey{repositoryName: "acme/widgets", pullRequestNumber: 42, filePath: "main.go", width: 80}] = reviewDiffRenderCacheEntry{}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch(updatePullRequestBranchActionTitle, matchingActionsPopupIndexes(subject.currentActionsPopupActions(), updatePullRequestBranchActionTitle))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)

	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued update-branch mutation, actual %d", len(asyncRunner.runs))
	}
	asyncRunner.runs[0]()

	if !reflect.DeepEqual(loader.updateBranchCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected update branch calls %v, actual %v", []string{"acme/widgets#42"}, loader.updateBranchCalls)
	}
	then_currentViewNameIs(t, gui, viewDetailName)
	then_statusLineContains(t, gui, pullRequestBranchUpdatedSuccessMessage)
	detailResult, ok := subject.pullRequestDetailCache["acme/widgets#42"]
	if !ok {
		t.Fatal("expected the pull request detail cache to stay warm after updating the branch")
	}
	if detailResult.detail.OutOfDateWithBase {
		t.Fatal("expected the optimistic detail cache to clear the out-of-date flag")
	}
	if !detailResult.needsRefresh {
		t.Fatal("expected the detail cache to queue a background refresh after updating the branch")
	}
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if strings.Contains(detailView.Buffer(), "Out of date with base branch") {
		t.Fatalf("expected the optimistic detail view to drop the out-of-date line, actual %q", detailView.Buffer())
	}
	if _, ok := subject.pullRequestDiffCache["acme/widgets#42"]; ok {
		t.Fatal("expected the cached pull-request diff to be invalidated after updating the branch")
	}
	if len(subject.reviewDiffRenderCache) != 0 {
		t.Fatalf("expected the review diff render cache to be cleared, actual %d entries", len(subject.reviewDiffRenderCache))
	}
	if len(asyncRunner.runs) != 2 {
		t.Fatalf("expected one queued update-branch mutation plus one queued detail refresh, actual %d", len(asyncRunner.runs))
	}
}

func TestActionsPopup_GivenAnEnableAutoMergeMutation_WhenExecuting_ThenItKeepsThePopupOpenShowsTheStatusLineSpinnerAndDelaysTheGitHubCall(t *testing.T) {
	summary := given_pullRequestLifecycleSummary("OPEN", false)
	summary.StatusCheckRollupState = "PENDING"
	model := given_pullRequestLifecycleModel(summary)
	model.OpenDetail()
	loader := &fakePullRequestDetailLoader{details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": given_pullRequestLifecycleDetailWithPendingBuild("OPEN", false, false)}}
	subject := given_pullRequestCommentProgram(model, loader)
	asyncRunner := &capturingAsyncRunner{}
	subject.asyncRunner = asyncRunner
	subject.uiUpdater = immediateUIUpdater{}
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(given_pullRequestLifecycleDetailWithPendingBuild("OPEN", false, false))}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch(enablePullRequestAutoMergeActionTitle, matchingActionsPopupIndexes(subject.currentActionsPopupActions(), enablePullRequestAutoMergeActionTitle))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)

	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued enable-auto-merge mutation, actual %d", len(asyncRunner.runs))
	}
	if len(loader.enableAutoMergeCalls) != 0 {
		t.Fatalf("expected the enable-auto-merge call to wait for the queued run, actual %v", loader.enableAutoMergeCalls)
	}
	then_currentViewNameIs(t, gui, viewActionsPopupSearchName)
	_, actualErr = gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	then_statusLineContains(t, gui, string(loadingSpinnerFrames[0]))
	then_statusLineContains(t, gui, "Running `gh pr merge 42 -R acme/widgets --auto --squash`.")
}

func TestLayout_GivenAnEnableAutoMergeMutation_WhenRendering_ThenItTogglesTheOptimisticAutoMergeStateAndShowsFeedback(t *testing.T) {
	summary := given_pullRequestLifecycleSummary("OPEN", false)
	summary.StatusCheckRollupState = "PENDING"
	model := given_pullRequestLifecycleModel(summary)
	model.OpenDetail()
	loader := &fakePullRequestDetailLoader{details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": given_pullRequestLifecycleDetailWithPendingBuild("OPEN", false, false)}}
	subject := given_pullRequestCommentProgram(model, loader)
	asyncRunner := &capturingAsyncRunner{}
	subject.asyncRunner = asyncRunner
	subject.uiUpdater = immediateUIUpdater{}
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(given_pullRequestLifecycleDetailWithPendingBuild("OPEN", false, false))}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if strings.Contains(detailView.Buffer(), "Auto-merge enabled") {
		t.Fatalf("expected the detail buffer to omit %q before enabling auto-merge, actual %q", "Auto-merge enabled", detailView.Buffer())
	}
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch(enablePullRequestAutoMergeActionTitle, matchingActionsPopupIndexes(subject.currentActionsPopupActions(), enablePullRequestAutoMergeActionTitle))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)

	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued enable-auto-merge mutation, actual %d", len(asyncRunner.runs))
	}
	asyncRunner.runs[0]()

	if !reflect.DeepEqual(loader.enableAutoMergeCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected enable-auto-merge calls %v, actual %v", []string{"acme/widgets#42"}, loader.enableAutoMergeCalls)
	}
	then_currentViewNameIs(t, gui, viewDetailName)
	then_statusLineContains(t, gui, pullRequestAutoMergeEnabledSuccessMessage)
	detailView, actualErr = gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "Auto-merge enabled") {
		t.Fatalf("expected the detail buffer to contain %q after enabling auto-merge, actual %q", "Auto-merge enabled", detailView.Buffer())
	}

	selectedSummary, ok := subject.model.SelectedPullRequestSummary()
	if !ok || selectedSummary.AutoMergeRequest == nil {
		t.Fatalf("expected the selected summary to keep the optimistic auto-merge state, actual %+v", selectedSummary)
	}
	detailResult, ok := subject.pullRequestDetailCache["acme/widgets#42"]
	if !ok || detailResult.detail.AutoMergeRequest == nil {
		t.Fatalf("expected the detail cache to keep the optimistic auto-merge state, actual %+v", detailResult.detail)
	}

	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if !strings.Contains(popupView.Buffer(), disablePullRequestAutoMergeActionTitle) {
		t.Fatalf("expected the reopened popup to contain %q, actual %q", disablePullRequestAutoMergeActionTitle, popupView.Buffer())
	}
}

func TestLayout_GivenADisableAutoMergeMutation_WhenRendering_ThenItClearsTheOptimisticAutoMergeStateAndShowsFeedback(t *testing.T) {
	summary := given_pullRequestLifecycleSummary("OPEN", false)
	summary.StatusCheckRollupState = "PENDING"
	summary.AutoMergeRequest = given_pullRequestAutoMergeRequest()
	model := given_pullRequestLifecycleModel(summary)
	model.OpenDetail()
	loader := &fakePullRequestDetailLoader{details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": given_pullRequestLifecycleDetailWithPendingBuild("OPEN", false, true)}}
	subject := given_pullRequestCommentProgram(model, loader)
	asyncRunner := &capturingAsyncRunner{}
	subject.asyncRunner = asyncRunner
	subject.uiUpdater = immediateUIUpdater{}
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(given_pullRequestLifecycleDetailWithPendingBuild("OPEN", false, true))}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "Auto-merge enabled") {
		t.Fatalf("expected the detail buffer to contain %q before disabling auto-merge, actual %q", "Auto-merge enabled", detailView.Buffer())
	}
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch(disablePullRequestAutoMergeActionTitle, matchingActionsPopupIndexes(subject.currentActionsPopupActions(), disablePullRequestAutoMergeActionTitle))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)

	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued disable-auto-merge mutation, actual %d", len(asyncRunner.runs))
	}
	asyncRunner.runs[0]()

	if !reflect.DeepEqual(loader.disableAutoMergeCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected disable-auto-merge calls %v, actual %v", []string{"acme/widgets#42"}, loader.disableAutoMergeCalls)
	}
	then_currentViewNameIs(t, gui, viewDetailName)
	then_statusLineContains(t, gui, pullRequestAutoMergeDisabledSuccessMessage)
	detailView, actualErr = gui.View(viewDetailName)
	then_noError(t, actualErr)
	if strings.Contains(detailView.Buffer(), "Auto-merge enabled") {
		t.Fatalf("expected the detail buffer to omit %q after disabling auto-merge, actual %q", "Auto-merge enabled", detailView.Buffer())
	}

	selectedSummary, ok := subject.model.SelectedPullRequestSummary()
	if !ok || selectedSummary.AutoMergeRequest != nil {
		t.Fatalf("expected the selected summary to clear the optimistic auto-merge state, actual %+v", selectedSummary)
	}
	detailResult, ok := subject.pullRequestDetailCache["acme/widgets#42"]
	if !ok || detailResult.detail.AutoMergeRequest != nil {
		t.Fatalf("expected the detail cache to clear the optimistic auto-merge state, actual %+v", detailResult.detail)
	}

	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if !strings.Contains(popupView.Buffer(), enablePullRequestAutoMergeActionTitle) {
		t.Fatalf("expected the reopened popup to contain %q, actual %q", enablePullRequestAutoMergeActionTitle, popupView.Buffer())
	}
}

func TestActionsPopup_GivenClosePullRequestFailure_WhenExecuting_ThenItKeepsTheUIStableAndShowsTheGitHubError(t *testing.T) {
	loader := &fakePullRequestDetailLoader{closePullRequestErr: errors.New("GitHub rejected the close")}
	model := given_pullRequestLifecycleModel(given_pullRequestLifecycleSummary("OPEN", false))
	model.OpenDetail()
	subject := given_pullRequestCommentProgram(model, loader)
	asyncRunner := &capturingAsyncRunner{}
	subject.asyncRunner = asyncRunner
	subject.uiUpdater = immediateUIUpdater{}
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(given_pullRequestLifecycleDetail("OPEN", false))}
	subject.pullRequestDiffCache["acme/widgets#42"] = pullRequestDiffResult{data: buildReviewDiffData(given_reviewSessionPullRequestDiff())}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch(closePullRequestActionTitle, matchingActionsPopupIndexes(subject.currentActionsPopupActions(), closePullRequestActionTitle))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)

	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)
	given_runQueuedAsync(t, asyncRunner, 0)

	then_currentViewNameIs(t, gui, viewActionsPopupSearchName)
	if !reflect.DeepEqual(loader.closePullRequestCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected close pull request calls %v, actual %v", []string{"acme/widgets#42"}, loader.closePullRequestCalls)
	}
	if subject.actionsPopupWidget.errorMessage != "" {
		t.Fatalf("expected popup error message to stay empty, actual %q", subject.actionsPopupWidget.errorMessage)
	}
	then_transientErrorPopupContains(t, gui, "GitHub rejected the close")
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "OPEN") {
		t.Fatalf("expected the detail buffer to keep %q after the failure, actual %q", "OPEN", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), "CLOSED") {
		t.Fatalf("expected the detail buffer to stay open after the failure, actual %q", detailView.Buffer())
	}
	if _, ok := subject.pullRequestDiffCache["acme/widgets#42"]; !ok {
		t.Fatal("expected the cached pull-request diff to stay intact after the failed close")
	}
}

func TestActionsPopup_GivenUpdateBranchFailure_WhenExecuting_ThenItKeepsTheUIStableAndShowsTheGitHubError(t *testing.T) {
	loader := &fakePullRequestDetailLoader{updateBranchErr: errors.New("GitHub rejected the branch update")}
	model := given_pullRequestLifecycleModel(given_pullRequestLifecycleSummary("OPEN", false))
	model.OpenDetail()
	subject := given_pullRequestCommentProgram(model, loader)
	asyncRunner := &capturingAsyncRunner{}
	subject.asyncRunner = asyncRunner
	subject.uiUpdater = immediateUIUpdater{}
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(given_pullRequestOutOfDateLifecycleDetail("OPEN", false))}
	subject.pullRequestDiffCache["acme/widgets#42"] = pullRequestDiffResult{data: buildReviewDiffData(given_reviewSessionPullRequestDiff())}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch(updatePullRequestBranchActionTitle, matchingActionsPopupIndexes(subject.currentActionsPopupActions(), updatePullRequestBranchActionTitle))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)

	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)
	given_runQueuedAsync(t, asyncRunner, 0)

	then_currentViewNameIs(t, gui, viewActionsPopupSearchName)
	if !reflect.DeepEqual(loader.updateBranchCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected update branch calls %v, actual %v", []string{"acme/widgets#42"}, loader.updateBranchCalls)
	}
	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if strings.Contains(popupView.Title, "GitHub rejected the branch update") {
		t.Fatalf("expected the popup title to hide %q, actual %q", "GitHub rejected the branch update", popupView.Title)
	}
	then_transientErrorPopupContains(t, gui, "GitHub rejected the branch update")
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "Out of date with base branch") {
		t.Fatalf("expected the detail to keep the out-of-date line after the failure, actual %q", detailView.Buffer())
	}
	if _, ok := subject.pullRequestDiffCache["acme/widgets#42"]; !ok {
		t.Fatal("expected the cached pull-request diff to stay intact after the failed branch update")
	}
}

func TestLayout_GivenAReopenClosedDraftPullRequestMutation_WhenRendering_ThenTheUpdatedDraftStateFeedbackAndCacheInvalidationAreVisible(t *testing.T) {
	summary := given_pullRequestLifecycleSummary("CLOSED", true)
	model := given_pullRequestLifecycleModel(summary)
	model.OpenDetail()
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": given_pullRequestLifecycleDetail("CLOSED", true),
		},
	}
	subject := given_pullRequestCommentProgram(model, loader)
	asyncRunner := &capturingAsyncRunner{}
	subject.asyncRunner = asyncRunner
	subject.uiUpdater = immediateUIUpdater{}
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(given_pullRequestLifecycleDetail("CLOSED", true))}
	subject.pullRequestDiffCache["acme/widgets#42"] = pullRequestDiffResult{data: buildReviewDiffData(given_reviewSessionPullRequestDiff())}
	subject.reviewDiffRenderCache[reviewDiffRenderCacheKey{repositoryName: "acme/widgets", pullRequestNumber: 42, filePath: "main.go", width: 80}] = reviewDiffRenderCacheEntry{}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch(reopenPullRequestActionTitle, matchingActionsPopupIndexes(subject.currentActionsPopupActions(), reopenPullRequestActionTitle))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)

	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued reopen mutation, actual %d", len(asyncRunner.runs))
	}
	asyncRunner.runs[0]()

	if !reflect.DeepEqual(loader.reopenPullRequestCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected reopen pull request calls %v, actual %v", []string{"acme/widgets#42"}, loader.reopenPullRequestCalls)
	}
	then_currentViewNameIs(t, gui, viewDetailName)
	then_statusLineContains(t, gui, pullRequestReopenedSuccessMessage)
	rows := subject.model.PullRequestRows(MyPullRequestsTab)
	if len(rows) != 1 || rows[0].Summary == nil || rows[0].Summary.State != "OPEN" || !rows[0].Summary.IsDraft {
		t.Fatalf("expected the visible pull request summary to be reopened as draft, actual %+v", rows)
	}
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "DRAFT") {
		t.Fatalf("expected the detail buffer to contain %q, actual %q", "DRAFT", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), "CLOSED") {
		t.Fatalf("expected the detail buffer to drop %q, actual %q", "CLOSED", detailView.Buffer())
	}
	if _, ok := subject.pullRequestDiffCache["acme/widgets#42"]; ok {
		t.Fatal("expected the cached pull-request diff to be invalidated after reopening the pull request")
	}
	if len(subject.reviewDiffRenderCache) != 0 {
		t.Fatalf("expected the review diff render cache to be cleared, actual %d entries", len(subject.reviewDiffRenderCache))
	}
	if len(asyncRunner.runs) != 2 {
		t.Fatalf("expected one queued reopen mutation plus one queued detail refresh, actual %d", len(asyncRunner.runs))
	}
}

func TestActionsPopup_GivenReopenPullRequestFailure_WhenExecuting_ThenItKeepsTheUIStableAndShowsTheGitHubError(t *testing.T) {
	loader := &fakePullRequestDetailLoader{reopenPullRequestErr: errors.New("GitHub rejected the reopen")}
	model := given_pullRequestLifecycleModel(given_pullRequestLifecycleSummary("CLOSED", true))
	model.OpenDetail()
	subject := given_pullRequestCommentProgram(model, loader)
	asyncRunner := &capturingAsyncRunner{}
	subject.asyncRunner = asyncRunner
	subject.uiUpdater = immediateUIUpdater{}
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(given_pullRequestLifecycleDetail("CLOSED", true))}
	subject.pullRequestDiffCache["acme/widgets#42"] = pullRequestDiffResult{data: buildReviewDiffData(given_reviewSessionPullRequestDiff())}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch(reopenPullRequestActionTitle, matchingActionsPopupIndexes(subject.currentActionsPopupActions(), reopenPullRequestActionTitle))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)

	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)
	given_runQueuedAsync(t, asyncRunner, 0)

	then_currentViewNameIs(t, gui, viewActionsPopupSearchName)
	if !reflect.DeepEqual(loader.reopenPullRequestCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected reopen pull request calls %v, actual %v", []string{"acme/widgets#42"}, loader.reopenPullRequestCalls)
	}
	if subject.actionsPopupWidget.errorMessage != "" {
		t.Fatalf("expected popup error message to stay empty, actual %q", subject.actionsPopupWidget.errorMessage)
	}
	then_transientErrorPopupContains(t, gui, "GitHub rejected the reopen")
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "CLOSED") {
		t.Fatalf("expected the detail buffer to keep %q after the failure, actual %q", "CLOSED", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), "OPEN") || strings.Contains(detailView.Buffer(), "DRAFT") {
		t.Fatalf("expected the detail buffer to stay closed after the failure, actual %q", detailView.Buffer())
	}
	if _, ok := subject.pullRequestDiffCache["acme/widgets#42"]; !ok {
		t.Fatal("expected the cached pull-request diff to stay intact after the failed reopen")
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
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)

	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	then_currentViewNameIs(t, gui, viewActionsPopupSearchName)
	if len(loader.squashMergeCalls) != 0 {
		t.Fatalf("expected no squash-merge calls before confirmation, actual %v", loader.squashMergeCalls)
	}
	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if !strings.Contains(popupView.Title, squashMergePullRequestConfirmationPromptMessage) {
		t.Fatalf("expected the popup title to contain %q, actual %q", squashMergePullRequestConfirmationPromptMessage, popupView.Title)
	}
}

func TestActionsPopup_GivenAConfirmedSquashMerge_WhenTheMutationIsQueued_ThenItClosesThePopupAndShowsTheLoadingSpinner(t *testing.T) {
	summary := given_pullRequestLifecycleSummary("OPEN", false)
	model := given_pullRequestLifecycleModel(summary)
	model.OpenDetail()
	loader := &fakePullRequestDetailLoader{details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": given_pullRequestLifecycleDetail("OPEN", false)}}
	subject := given_pullRequestCommentProgram(model, loader)
	asyncRunner := &capturingAsyncRunner{}
	subject.asyncRunner = asyncRunner
	subject.uiUpdater = immediateUIUpdater{}
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(given_pullRequestLifecycleDetail("OPEN", false))}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("squash", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "squash"))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	then_currentViewNameIs(t, gui, viewDetailName)
	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued squash-merge run, actual %d", len(asyncRunner.runs))
	}
	then_statusLineContains(t, gui, string(loadingSpinnerFrames[0]))
	then_statusLineContains(t, gui, "Running `gh pr merge 42 -R acme/widgets --squash`.")
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "OPEN") {
		t.Fatalf("expected the detail buffer to keep %q while the squash merge is in flight, actual %q", "OPEN", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), "MERGED") {
		t.Fatalf("expected the detail buffer to stay unmerged while the squash merge is in flight, actual %q", detailView.Buffer())
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
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(given_pullRequestLifecycleDetail("OPEN", false))}
	subject.pullRequestDiffCache["acme/widgets#42"] = pullRequestDiffResult{data: buildReviewDiffData(given_reviewSessionPullRequestDiff())}
	subject.reviewDiffRenderCache[reviewDiffRenderCacheKey{repositoryName: "acme/widgets", pullRequestNumber: 42, filePath: "main.go", width: 80}] = reviewDiffRenderCacheEntry{}
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
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued squash-merge run, actual %d", len(asyncRunner.runs))
	}
	if len(loader.squashMergeCalls) != 0 {
		t.Fatalf("expected the squash merge call to wait for the queued run, actual %v", loader.squashMergeCalls)
	}

	asyncRunner.runs[0]()

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
}

func TestActionsPopup_GivenAConfirmedSquashMergeFailure_WhenExecuting_ThenItKeepsTheUIStableAndShowsTheGitHubError(t *testing.T) {
	loader := &fakePullRequestDetailLoader{squashMergeErr: errors.New("GitHub rejected the squash merge")}
	model := given_pullRequestLifecycleModel(given_pullRequestLifecycleSummary("OPEN", false))
	model.OpenDetail()
	subject := given_pullRequestCommentProgram(model, loader)
	asyncRunner := &capturingAsyncRunner{}
	subject.asyncRunner = asyncRunner
	subject.uiUpdater = immediateUIUpdater{}
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(given_pullRequestLifecycleDetail("OPEN", false))}
	subject.pullRequestDiffCache["acme/widgets#42"] = pullRequestDiffResult{data: buildReviewDiffData(given_reviewSessionPullRequestDiff())}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("squash", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "squash"))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	then_currentViewNameIs(t, gui, viewDetailName)
	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued squash-merge run, actual %d", len(asyncRunner.runs))
	}
	if len(loader.squashMergeCalls) != 0 {
		t.Fatalf("expected the squash merge call to wait for the queued run, actual %v", loader.squashMergeCalls)
	}

	asyncRunner.runs[0]()

	if !reflect.DeepEqual(loader.squashMergeCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected squash-merge calls %v, actual %v", []string{"acme/widgets#42"}, loader.squashMergeCalls)
	}
	then_statusLineDoesNotContain(t, gui, "GitHub rejected the squash merge")
	then_transientErrorPopupContains(t, gui, "GitHub rejected the squash merge")
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

func given_pullRequestAutoMergeRequest() *githubcli.PullRequestAutoMergeRequest {
	return &githubcli.PullRequestAutoMergeRequest{EnabledAt: "2026-05-20T10:00:00Z"}
}

func given_pullRequestLifecycleDetailWithPendingBuild(state string, isDraft bool, autoMergeEnabled bool) githubcli.PullRequestDetail {
	detail := given_pullRequestLifecycleDetail(state, isDraft)
	detail.StatusCheckRollup = []githubcli.PullRequestStatusCheck{{Name: "CI", Status: "IN_PROGRESS"}}
	if autoMergeEnabled {
		detail.AutoMergeRequest = given_pullRequestAutoMergeRequest()
	}
	return detail
}

func given_pullRequestOutOfDateLifecycleDetail(state string, isDraft bool) githubcli.PullRequestDetail {
	detail := given_pullRequestLifecycleDetail(state, isDraft)
	detail.OutOfDateWithBase = true
	detail.MergeStateStatus = "BLOCKED"
	return detail
}
