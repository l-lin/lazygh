package tui

import (
	"testing"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func TestUpdate_GivenMsgOpenPullRequestInBrowserRequested_WhenApplying_ThenItQueuesAnAsyncBrowserMutationCmd(t *testing.T) {
	loader := &fakePullRequestDetailLoader{}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)

	actual := Update(subject, MsgOpenPullRequestInBrowserRequested{Target: pullRequestActionTarget{repository: "acme/widgets", number: 42}})

	if len(actual) != 1 {
		t.Fatalf("expected one queued command, actual %d", len(actual))
	}
	command := given_actionsPopupAsyncCommand(t, actual)
	request, ok := command.request.(openPullRequestInBrowserPopupRequest)
	if !ok {
		t.Fatalf("expected an openPullRequestInBrowserPopupRequest, actual %T", command.request)
	}
	if !request.asyncRequested() {
		t.Fatal("expected the browser mutation command to run asynchronously")
	}
	if actualCommand := request.statusCommand(); actualCommand != openPullRequestInBrowserCommand("acme/widgets", 42) {
		t.Fatalf("expected browser command %q, actual %q", openPullRequestInBrowserCommand("acme/widgets", 42), actualCommand)
	}

	actualSuccess, actualErr := request.run(subject)
	then_noError(t, actualErr)
	if len(loader.openBrowserCalls) != 1 || loader.openBrowserCalls[0] != "acme/widgets#42" {
		t.Fatalf("expected open browser calls %v, actual %v", []string{"acme/widgets#42"}, loader.openBrowserCalls)
	}
	feedbackSuccess, ok := actualSuccess.(actionsPopupAsyncFeedbackSuccess)
	if !ok {
		t.Fatalf("expected feedback success, actual %T", actualSuccess)
	}
	if actualMessage := feedbackSuccess.Message; actualMessage != pullRequestBrowserOpenSuccessMessage {
		t.Fatalf("expected success message %q, actual %q", pullRequestBrowserOpenSuccessMessage, actualMessage)
	}
}

func TestUpdate_GivenMsgApprovePullRequestRequested_WhenApplying_ThenItQueuesAnAsyncInvalidateDiffMutationCmd(t *testing.T) {
	loader := &fakePullRequestDetailLoader{}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)

	actual := Update(subject, MsgApprovePullRequestRequested{Target: pullRequestActionTarget{repository: "acme/widgets", number: 42}})

	if len(actual) != 1 {
		t.Fatalf("expected one queued command, actual %d", len(actual))
	}
	command := given_actionsPopupAsyncCommand(t, actual)
	request, ok := command.request.(approvePullRequestPopupRequest)
	if !ok {
		t.Fatalf("expected an approvePullRequestPopupRequest, actual %T", command.request)
	}
	if !request.asyncRequested() {
		t.Fatal("expected the review approval command to run asynchronously")
	}
	if actualCommand := request.statusCommand(); actualCommand != approvePullRequestCommand("acme/widgets", 42) {
		t.Fatalf("expected approve command %q, actual %q", approvePullRequestCommand("acme/widgets", 42), actualCommand)
	}

	actualSuccess, actualErr := request.run(subject)
	then_noError(t, actualErr)
	if len(loader.approveCalls) != 1 || loader.approveCalls[0] != "acme/widgets#42" {
		t.Fatalf("expected approve calls %v, actual %v", []string{"acme/widgets#42"}, loader.approveCalls)
	}
	invalidateSuccess, ok := actualSuccess.(actionsPopupAsyncInvalidatePullRequestSuccess)
	if !ok {
		t.Fatalf("expected invalidate success, actual %T", actualSuccess)
	}
	if invalidateSuccess.Repository != "acme/widgets" || invalidateSuccess.Number != 42 || !invalidateSuccess.InvalidateDiff || invalidateSuccess.Message != pullRequestReviewSuccessMessage {
		t.Fatalf("expected invalidate success %+v, actual %+v", actionsPopupAsyncInvalidatePullRequestSuccess{Repository: "acme/widgets", Number: 42, InvalidateDiff: true, Message: pullRequestReviewSuccessMessage}, invalidateSuccess)
	}
}

func TestUpdate_GivenMsgReRequestPullRequestReviewRequested_WhenApplying_ThenItQueuesAnAsyncInvalidateDetailMutationCmd(t *testing.T) {
	loader := given_pullRequestReviewerLoader()
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)

	actual := Update(subject, MsgReRequestPullRequestReviewRequested{Target: pullRequestReviewerRequestTarget{repository: "acme/widgets", number: 42, reviewerLogin: "reviewer-approved"}})

	if len(actual) != 1 {
		t.Fatalf("expected one queued command, actual %d", len(actual))
	}
	command := given_actionsPopupAsyncCommand(t, actual)
	request, ok := command.request.(reRequestPullRequestReviewPopupRequest)
	if !ok {
		t.Fatalf("expected a reRequestPullRequestReviewPopupRequest, actual %T", command.request)
	}
	if !request.asyncRequested() {
		t.Fatal("expected the reviewer re-request command to run asynchronously")
	}
	if actualCommand := request.statusCommand(); actualCommand != requestPullRequestReviewerCommand("acme/widgets", 42, "reviewer-approved") {
		t.Fatalf("expected reviewer command %q, actual %q", requestPullRequestReviewerCommand("acme/widgets", 42, "reviewer-approved"), actualCommand)
	}

	actualSuccess, actualErr := request.run(subject)
	then_noError(t, actualErr)
	if len(loader.requestReviewerCalls) != 1 || loader.requestReviewerCalls[0] != "acme/widgets#42" {
		t.Fatalf("expected reviewer request calls %v, actual %v", []string{"acme/widgets#42"}, loader.requestReviewerCalls)
	}
	if len(loader.requestReviewerLogins) != 1 || loader.requestReviewerLogins[0] != "reviewer-approved" {
		t.Fatalf("expected reviewer request logins %v, actual %v", []string{"reviewer-approved"}, loader.requestReviewerLogins)
	}
	invalidateSuccess, ok := actualSuccess.(actionsPopupAsyncInvalidatePullRequestSuccess)
	if !ok {
		t.Fatalf("expected invalidate success, actual %T", actualSuccess)
	}
	if invalidateSuccess.Repository != "acme/widgets" || invalidateSuccess.Number != 42 || invalidateSuccess.InvalidateDiff || invalidateSuccess.Message != pullRequestReviewReRequestedSuccessMessage {
		t.Fatalf("expected invalidate success %+v, actual %+v", actionsPopupAsyncInvalidatePullRequestSuccess{Repository: "acme/widgets", Number: 42, Message: pullRequestReviewReRequestedSuccessMessage}, invalidateSuccess)
	}
}

func TestUpdate_GivenMsgPullRequestLifecycleMutationRequested_WhenApplyingMarkReadyForReview_ThenItQueuesAnAsyncOptimisticLifecycleMutationCmd(t *testing.T) {
	loader := &fakePullRequestDetailLoader{}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	summary := given_pullRequestMutationSummary("OPEN", true)

	actual := Update(subject, MsgPullRequestLifecycleMutationRequested{
		Kind:           pullRequestLifecycleMutationReadyForReview,
		Target:         pullRequestActionTarget{repository: "acme/widgets", number: 42},
		Summary:        summary,
		State:          "OPEN",
		IsDraft:        false,
		SuccessMessage: pullRequestMarkedReadyForReviewSuccessMessage,
	})

	if len(actual) != 1 {
		t.Fatalf("expected one queued command, actual %d", len(actual))
	}
	command := given_actionsPopupAsyncCommand(t, actual)
	request, ok := command.request.(pullRequestLifecycleMutationPopupRequest)
	if !ok {
		t.Fatalf("expected a pullRequestLifecycleMutationPopupRequest, actual %T", command.request)
	}
	if !request.asyncRequested() {
		t.Fatal("expected the lifecycle mutation command to run asynchronously")
	}
	if actualCommand := request.statusCommand(); actualCommand != pullRequestReadyCommand("acme/widgets", 42, false) {
		t.Fatalf("expected lifecycle command %q, actual %q", pullRequestReadyCommand("acme/widgets", 42, false), actualCommand)
	}

	actualSuccess, actualErr := request.run(subject)
	then_noError(t, actualErr)
	if len(loader.markReadyForReviewCalls) != 1 || loader.markReadyForReviewCalls[0] != "acme/widgets#42" {
		t.Fatalf("expected ready-for-review calls %v, actual %v", []string{"acme/widgets#42"}, loader.markReadyForReviewCalls)
	}
	lifecycleSuccess, ok := actualSuccess.(actionsPopupAsyncPullRequestLifecycleSuccess)
	if !ok {
		t.Fatalf("expected lifecycle success, actual %T", actualSuccess)
	}
	if !samePullRequestIdentity(lifecycleSuccess.Summary, summary) || lifecycleSuccess.State != "OPEN" || lifecycleSuccess.IsDraft || lifecycleSuccess.Message != pullRequestMarkedReadyForReviewSuccessMessage {
		t.Fatalf("expected lifecycle success for %v with OPEN/non-draft/%q, actual %+v", summary, pullRequestMarkedReadyForReviewSuccessMessage, lifecycleSuccess)
	}
}

func TestUpdate_GivenMsgPullRequestAutoMergeMutationRequested_WhenApplyingEnableAutoMerge_ThenItQueuesAnAsyncOptimisticAutoMergeMutationCmd(t *testing.T) {
	loader := &fakePullRequestDetailLoader{}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	summary := given_pullRequestMutationSummary("OPEN", false)

	actual := Update(subject, MsgPullRequestAutoMergeMutationRequested{
		Kind:           pullRequestAutoMergeMutationEnable,
		Target:         pullRequestActionTarget{repository: "acme/widgets", number: 42},
		Summary:        summary,
		Enabled:        true,
		SuccessMessage: pullRequestAutoMergeEnabledSuccessMessage,
	})

	if len(actual) != 1 {
		t.Fatalf("expected one queued command, actual %d", len(actual))
	}
	command := given_actionsPopupAsyncCommand(t, actual)
	request, ok := command.request.(pullRequestAutoMergeMutationPopupRequest)
	if !ok {
		t.Fatalf("expected a pullRequestAutoMergeMutationPopupRequest, actual %T", command.request)
	}
	if !request.asyncRequested() {
		t.Fatal("expected the auto-merge mutation command to run asynchronously")
	}
	if actualCommand := request.statusCommand(); actualCommand != enablePullRequestAutoMergeCommand("acme/widgets", 42) {
		t.Fatalf("expected auto-merge command %q, actual %q", enablePullRequestAutoMergeCommand("acme/widgets", 42), actualCommand)
	}

	actualSuccess, actualErr := request.run(subject)
	then_noError(t, actualErr)
	if len(loader.enableAutoMergeCalls) != 1 || loader.enableAutoMergeCalls[0] != "acme/widgets#42" {
		t.Fatalf("expected enable auto-merge calls %v, actual %v", []string{"acme/widgets#42"}, loader.enableAutoMergeCalls)
	}
	autoMergeSuccess, ok := actualSuccess.(actionsPopupAsyncPullRequestAutoMergeSuccess)
	if !ok {
		t.Fatalf("expected auto-merge success, actual %T", actualSuccess)
	}
	if !samePullRequestIdentity(autoMergeSuccess.Summary, summary) || !autoMergeSuccess.Enabled || autoMergeSuccess.Message != pullRequestAutoMergeEnabledSuccessMessage {
		t.Fatalf("expected auto-merge success for %v with enabled/%q, actual %+v", summary, pullRequestAutoMergeEnabledSuccessMessage, autoMergeSuccess)
	}
}

func TestUpdate_GivenMsgPullRequestBranchUpdateRequested_WhenApplying_ThenItQueuesAnAsyncOptimisticBranchUpdateMutationCmd(t *testing.T) {
	loader := &fakePullRequestDetailLoader{}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	summary := given_pullRequestMutationSummary("OPEN", false)

	actual := Update(subject, MsgPullRequestBranchUpdateRequested{Target: pullRequestActionTarget{repository: "acme/widgets", number: 42}, Summary: summary})

	if len(actual) != 1 {
		t.Fatalf("expected one queued command, actual %d", len(actual))
	}
	command := given_actionsPopupAsyncCommand(t, actual)
	request, ok := command.request.(pullRequestBranchUpdatePopupRequest)
	if !ok {
		t.Fatalf("expected a pullRequestBranchUpdatePopupRequest, actual %T", command.request)
	}
	if !request.asyncRequested() {
		t.Fatal("expected the branch update command to run asynchronously")
	}
	if actualCommand := request.statusCommand(); actualCommand != updatePullRequestBranchCommand("acme/widgets", 42) {
		t.Fatalf("expected branch update command %q, actual %q", updatePullRequestBranchCommand("acme/widgets", 42), actualCommand)
	}

	actualSuccess, actualErr := request.run(subject)
	then_noError(t, actualErr)
	if len(loader.updateBranchCalls) != 1 || loader.updateBranchCalls[0] != "acme/widgets#42" {
		t.Fatalf("expected branch update calls %v, actual %v", []string{"acme/widgets#42"}, loader.updateBranchCalls)
	}
	branchSuccess, ok := actualSuccess.(actionsPopupAsyncPullRequestBranchUpdateSuccess)
	if !ok {
		t.Fatalf("expected branch update success, actual %T", actualSuccess)
	}
	if !samePullRequestIdentity(branchSuccess.Summary, summary) || branchSuccess.Message != pullRequestBranchUpdatedSuccessMessage {
		t.Fatalf("expected branch update success for %v with %q, actual %+v", summary, pullRequestBranchUpdatedSuccessMessage, branchSuccess)
	}
}

func given_actionsPopupAsyncCommand(t *testing.T, commands []Cmd) actionsPopupAsyncCmd {
	t.Helper()

	if len(commands) != 1 {
		t.Fatalf("expected one queued command, actual %d", len(commands))
	}

	command, ok := commands[0].(actionsPopupAsyncCmd)
	if !ok {
		t.Fatalf("expected an actionsPopupAsyncCmd, actual %T", commands[0])
	}
	if command.request == nil {
		t.Fatal("expected the actions popup async command to include a typed request")
	}
	return command
}

func given_pullRequestMutationSummary(state string, isDraft bool) githubdomain.PullRequest {
	return githubdomain.PullRequest{
		Title:      "Lifecycle PR",
		Number:     42,
		Repository: githubdomain.Repository{NameWithOwner: "acme/widgets"},
		URL:        "https://github.com/acme/widgets/pull/42",
		Body:       "Lifecycle body",
		State:      state,
		IsDraft:    isDraft,
	}
}
