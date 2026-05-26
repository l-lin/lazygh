package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

func TestRefactorGuard_GivenProductionFiles_WhenScanning_ThenOnlyModelAndUpdateFilesWriteProgramModelFields(t *testing.T) {
	assignmentPattern := regexp.MustCompile(`program\.model\.[A-Za-z0-9_]+(?:\[[^\]]+\])?\s*=\s*[^=]`)
	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", assignmentPattern, func(path string) bool {
		base := filepath.Base(path)
		return strings.HasSuffix(base, ".go") && !strings.HasSuffix(base, "_test.go")
	})

	remainingMatches := make([]string, 0, len(actualMatches))
	for _, match := range actualMatches {
		base := filepath.Base(strings.Split(match, ":")[0])
		if strings.HasPrefix(base, "model") || strings.HasPrefix(base, "update") {
			continue
		}
		remainingMatches = append(remainingMatches, match)
	}
	if len(remainingMatches) != 0 {
		t.Fatalf("expected reducer/model transitions to own all program.model writes, actual %v", remainingMatches)
	}
}

func TestRefactorGuard_GivenRenderLayerFiles_WhenScanning_ThenTheyDoNotTriggerShellEffects(t *testing.T) {
	forbiddenPatterns := []*regexp.Regexp{
		regexp.MustCompile(`maybeLoad`),
		regexp.MustCompile(`uiUpdater\.Apply\(`),
		regexp.MustCompile(`detailImageManager\.Sync\(`),
		regexp.MustCompile(`WriteGraphicsCommand\(`),
	}

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", regexp.MustCompile(strings.Join([]string{
		forbiddenPatterns[0].String(),
		forbiddenPatterns[1].String(),
		forbiddenPatterns[2].String(),
		forbiddenPatterns[3].String(),
	}, "|")), given_isRenderLayerFile)
	if len(actualMatches) != 0 {
		t.Fatalf("expected render-layer files to stay free of shell effects, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenRenderLayerFiles_WhenScanning_ThenTheyDoNotMutateSelectorCachesOrConsumeOneShotState(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`program\.(?:cache|invalidate)[A-Za-z0-9_]+\(`,
		`program\.(?:pullRequestDetailDocumentForKey|pullRequestConversationDocumentForKey|pullRequestChangesRenderedRowsForKey|cachedReviewDiffRenderEntry|storeReviewDiffRenderEntry|consumePendingListViewportPlacement)\(`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, given_isRenderLayerFile)
	if len(actualMatches) != 0 {
		t.Fatalf("expected render-layer files to stay free of selector-cache mutation and one-shot consumption, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenProductionFiles_WhenScanning_ThenOnlyDispatchUsesUiUpdaterApply(t *testing.T) {
	allowedFiles := map[string]bool{
		"dispatch.go": true,
	}

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", regexp.MustCompile(`uiUpdater\.Apply\(`), func(path string) bool {
		base := filepath.Base(path)
		return strings.HasSuffix(base, ".go") && !strings.HasSuffix(base, "_test.go")
	})

	remainingMatches := make([]string, 0, len(actualMatches))
	for _, match := range actualMatches {
		base := filepath.Base(strings.Split(match, ":")[0])
		if allowedFiles[base] {
			continue
		}
		remainingMatches = append(remainingMatches, match)
	}
	if len(remainingMatches) != 0 {
		t.Fatalf("expected uiUpdater.Apply to stay confined to dispatch.go, actual %v", remainingMatches)
	}
}

func TestRefactorGuard_GivenPhase1NavigationFiles_WhenScanning_ThenTheyDoNotMutateProgramModelOrCallDirectShellRefreshHelpers(t *testing.T) {
	phase1Files := map[string]bool{
		"program_navigation.go":         true,
		"program_navigation_support.go": true,
		"review_file_tree_search.go":    true,
		"view_url.go":                   true,
	}
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`program\.model\.(?:Set|Open|Close|Select|Move|Update|Start|Cancel|Clear|Grow|Shrink|Focus|Blur|Submit|Advance|Cycle|Toggle|Reset|Remove|Add|Apply|Mark|Restore|Use)[A-Z][A-Za-z0-9_]*\(`,
		`program\.model\.adjustSelectionBy\(`,
		`syncCurrentView\(`,
		`refreshDetailView\(`,
		`reloadActivePullRequestsTab\(`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		return phase1Files[filepath.Base(path)]
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected phase 1 navigation surfaces to route transitions through reducer-owned helpers, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenPhase2ActionFiles_WhenScanning_ThenTheyDoNotMutateStateFromPopupActionSurfaces(t *testing.T) {
	phase2Files := map[string]bool{
		"cache_clear.go":                 true,
		"review_session.go":              true,
		"pull_request_custom_search.go":  true,
		"pull_request_assignee.go":       true,
		"reaction_picker.go":             true,
		"theme_picker.go":                true,
		"pull_request_refresh.go":        true,
		"pull_request_edit.go":           true,
		"pending_pull_request_review.go": true,
	}
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`program\.model\.(?:Set|Open|Close|Select|Move|Update|Start|Cancel|Clear|Grow|Shrink|Focus|Blur|Submit|Advance|Cycle|Toggle|Reset|Remove|Add|Apply|Mark|Restore|Use)[A-Z][A-Za-z0-9_]*\(`,
		`program\.actionsPopupWidget\.[A-Za-z0-9_]+\s*=`,
		`reloadActivePullRequestsTab\(`,
		`setFeedback\(`,
		`clearCachedData\(`,
		`openPullRequestReview\(`,
		`upsertPullRequestCustomSearch\(`,
		`openAssigneePicker\(`,
		`startAssigneePickerWarmup\(`,
		`toggleAssigneePickerSelection\(`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		return phase2Files[filepath.Base(path)]
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected phase 2 popup action files to route mutations through reducer-owned messages and commands, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenPhase1PopupAsyncFiles_WhenScanning_ThenTheyDoNotCallTheLegacyAsyncPopupBridge(t *testing.T) {
	phase1Files := map[string]bool{
		"pull_request_browser.go":     true,
		"pull_request_review.go":      true,
		"pull_request_reviewer.go":    true,
		"pull_request_stage_merge.go": true,
	}

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", regexp.MustCompile(`startActionsPopupAsyncGHCommand\(`), func(path string) bool {
		return phase1Files[filepath.Base(path)]
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected phase 1 popup async files to use reducer-owned request messages instead of the legacy async popup bridge, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenPhase2PopupFeedbackFiles_WhenScanning_ThenTheyDoNotOwnFeedbackReportingOrPopupCloseState(t *testing.T) {
	phase2Files := map[string]bool{
		"actions_popup_actions.go":          true,
		"comment_on_pr.go":                  true,
		"inline_comment_edit.go":            true,
		"inline_comment_reply.go":           true,
		"inline_comment_resolution.go":      true,
		"pull_request_comment_edit.go":      true,
		"pull_request_refresh_reporting.go": true,
		"pull_request_review.go":            true,
		"pull_request_stage_merge.go":       true,
		"reaction_remove.go":                true,
		"review_inline_comment.go":          true,
		"review_submit.go":                  true,
		"yank_motion.go":                    true,
	}
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`program\.(?:setFeedback|reportError)\(`,
		`actionsPopupActionResult\{closePopup:\s*true\}`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		return phase2Files[filepath.Base(path)]
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected phase 2 popup feedback files to route feedback, reporting, and popup close state through update-owned helpers or messages, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenProductionFiles_WhenScanning_ThenDirectFeedbackAndErrorReportingStayInUpdateFilesOrExplicitInteractionCommands(t *testing.T) {
	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", regexp.MustCompile(`program\.(?:setFeedback|reportError)\(`), func(path string) bool {
		base := filepath.Base(path)
		return strings.HasSuffix(base, ".go") && !strings.HasSuffix(base, "_test.go")
	})

	remainingMatches := make([]string, 0, len(actualMatches))
	for _, match := range actualMatches {
		base := filepath.Base(strings.Split(match, ":")[0])
		if strings.HasPrefix(base, "update") || base == "cmd_interaction.go" {
			continue
		}
		remainingMatches = append(remainingMatches, match)
	}
	if len(remainingMatches) != 0 {
		t.Fatalf("expected direct feedback and error reporting to stay in update files or explicit interaction commands, actual %v", remainingMatches)
	}
}

func TestRefactorGuard_GivenProductionFiles_WhenScanning_ThenNoLegacyAsyncPopupBridgeRemains(t *testing.T) {
	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", regexp.MustCompile(`startActionsPopupAsyncGHCommand\(`), func(path string) bool {
		base := filepath.Base(path)
		return strings.HasSuffix(base, ".go") && !strings.HasSuffix(base, "_test.go")
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected the legacy async popup bridge to be retired, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenProductionFiles_WhenScanning_ThenNoLegacyActionsPopupAsyncWorkCommandRemains(t *testing.T) {
	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", regexp.MustCompile(`actionsPopupAsyncWorkCmd|Work\s+func\(\*Program\)\s*\(actionsPopupAsyncSuccess, error\)`), func(path string) bool {
		base := filepath.Base(path)
		return strings.HasSuffix(base, ".go") && !strings.HasSuffix(base, "_test.go")
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected popup async work closures to be replaced by typed requests, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenProductionFiles_WhenScanning_ThenNoLegacyModalEditorSubmitCallbacksRemain(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`submit\s+func\(string\)\s+error`,
		`AfterSubmit\s+func\(\*Program\)\s+\[\]Cmd`,
		`afterSubmit\s+func\(\*gocui\.Gui\)`,
		`Submit:\s+func\(`,
		`AfterSubmit:\s+func\(`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		base := filepath.Base(path)
		return strings.HasSuffix(base, ".go") && !strings.HasSuffix(base, "_test.go")
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected modal editor submits to use typed requests and success handlers instead of legacy callbacks, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenProductionFiles_WhenScanning_ThenNoLegacyPopupFeatureWorkCallbacksRemain(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`Work\s+func\(\*Program\)\s+error`,
		`func \(program \*Program\) prepareStoryReview\(`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		base := filepath.Base(path)
		return strings.HasSuffix(base, ".go") && !strings.HasSuffix(base, "_test.go")
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected popup feature requests to use typed executors instead of legacy work callbacks, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenStartupReviewAndStoryUrlEntrypoints_WhenScanning_ThenTheyOnlyParseValidateAndDispatch(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`startReviewSession\(`,
		`applyPreparedStoryReview\(`,
		`layout\(program\.gui\)`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		base := filepath.Base(path)
		return base == "review_url.go" || base == "review_story_url.go"
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected startup review/story URL entrypoints to stay on parse-and-dispatch behavior, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenRemainingPopupActionFiles_WhenScanning_ThenTheyDoNotReturnLegacyClosePopupState(t *testing.T) {
	remainingPopupFiles := map[string]bool{
		"error_popup.go":            true,
		"modal_editor_lifecycle.go": true,
		"notification_actions.go":   true,
		"open_link.go":              true,
		"pull_request_build.go":     true,
		"refresh_active_view.go":    true,
		"review_story.go":           true,
	}

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", regexp.MustCompile(`actionsPopupActionResult\{closePopup:\s*true\}`), func(path string) bool {
		return remainingPopupFiles[filepath.Base(path)]
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected the remaining popup action files to close through reducer-owned messages instead of legacy action results, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenPhase1PopupResultFiles_WhenScanning_ThenTheyDoNotUseActionsPopupActionResultCompatibilityGlue(t *testing.T) {
	phase1Files := map[string]bool{
		"actions_popup_actions.go":  true,
		"error_popup.go":            true,
		"modal_editor_lifecycle.go": true,
		"notification_actions.go":   true,
		"open_link.go":              true,
		"pull_request_build.go":     true,
		"review_story.go":           true,
	}

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", regexp.MustCompile(`actionsPopupActionResult`), func(path string) bool {
		return phase1Files[filepath.Base(path)]
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected the low-risk phase 1 popup files to stop using actionsPopupActionResult compatibility glue, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenPhase2PopupResultFiles_WhenScanning_ThenTheyDoNotUseActionsPopupActionResultCompatibilityGlue(t *testing.T) {
	phase2Files := map[string]bool{
		"pending_pull_request_review.go": true,
		"pull_request_assignee.go":       true,
		"pull_request_browser.go":        true,
		"pull_request_refresh.go":        true,
		"pull_request_review.go":         true,
		"pull_request_reviewer.go":       true,
		"pull_request_stage_merge.go":    true,
		"reaction_picker.go":             true,
		"reaction_remove.go":             true,
		"review_session.go":              true,
		"theme_picker.go":                true,
	}

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", regexp.MustCompile(`actionsPopupActionResult`), func(path string) bool {
		return phase2Files[filepath.Base(path)]
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected the phase 2 popup feature files to stop using actionsPopupActionResult compatibility glue, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenProductionFiles_WhenScanning_ThenNoActionsPopupActionResultCompatibilityGlueRemains(t *testing.T) {
	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", regexp.MustCompile(`actionsPopupActionResult`), func(path string) bool {
		base := filepath.Base(path)
		return strings.HasSuffix(base, ".go") && !strings.HasSuffix(base, "_test.go")
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected actionsPopupActionResult compatibility glue to be fully retired, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenPhase1PopupEditorMutationFiles_WhenScanning_ThenTheyDoNotCallGitHubMutationPortsDirectly(t *testing.T) {
	phase1Files := map[string]bool{
		"comment_on_pr.go":             true,
		"inline_comment_edit.go":       true,
		"inline_comment_reply.go":      true,
		"inline_comment_resolution.go": true,
		"pull_request_comment_edit.go": true,
		"pull_request_edit.go":         true,
		"pull_request_review.go":       true,
		"pull_request_stage_merge.go":  true,
		"reaction_remove.go":           true,
		"review_inline_comment.go":     true,
		"review_submit.go":             true,
	}

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", regexp.MustCompile(`program\.(?:pullRequestMutations|reviewMutations|reactionMutations)\.[A-Za-z0-9_]+\(`), func(path string) bool {
		return phase1Files[filepath.Base(path)]
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected the phase 1 popup/editor mutation files to route GitHub mutations through update-owned commands, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenPhase2PopupFeatureFiles_WhenScanning_ThenTheyDoNotCallGitHubPortsDirectly(t *testing.T) {
	phase2Files := map[string]bool{
		"notification_actions.go": true,
		"review_story.go":         true,
	}

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", regexp.MustCompile(`program\.(?:notificationMutations|detailQueries|reviewMutations)\.[A-Za-z0-9_]+\(`), func(path string) bool {
		return phase2Files[filepath.Base(path)]
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected the phase 2 popup feature files to route GitHub queries and mutations through update-owned commands, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenProductionFiles_WhenScanning_ThenOnlyWorkflowCommandsCallGitHubPortsDirectly(t *testing.T) {
	allowedFiles := map[string]bool{
		"workflow_commands.go": true,
	}

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", regexp.MustCompile(`program\.(?:pullRequestMutations|reviewMutations|reactionMutations|notificationMutations|detailQueries|buildQueries)\.[A-Za-z0-9_]+\(`), func(path string) bool {
		base := filepath.Base(path)
		return strings.HasSuffix(base, ".go") && !strings.HasSuffix(base, "_test.go")
	})

	remainingMatches := make([]string, 0, len(actualMatches))
	for _, match := range actualMatches {
		base := filepath.Base(strings.Split(match, ":")[0])
		if allowedFiles[base] {
			continue
		}
		remainingMatches = append(remainingMatches, match)
	}
	if len(remainingMatches) != 0 {
		t.Fatalf("expected direct GitHub ports to stay confined to update-owned files and explicit command files, actual %v", remainingMatches)
	}
}

func TestRefactorGuard_GivenPullRequestCommandSurface_WhenScanning_ThenItDoesNotMutatePullRequestListLoadStateDirectly(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`program\.(?:myPullRequestsLoadStarted|requestedPullRequestsLoadStarted|myPullRequestsLoading|requestedPullRequestsLoading)\s*=`,
		`program\.(?:myPullRequestsCount|myPullRequestsCountKnown|requestedPullRequestsCount|requestedPullRequestsCountKnown)\s*=`,
		`program\.(?:additionalPullRequestsLoadStarted|additionalPullRequestsLoading|additionalPullRequestsCounts)\s*=`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		return filepath.Base(path) == "pull_request_commands.go"
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected pull_request_commands.go to stay on search descriptors and tab helpers instead of mutating pull request list load state, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenProductionFiles_WhenScanning_ThenOnlyUpdateFilesAndTheStoreHelperResetPullRequestListLoadState(t *testing.T) {
	allowedFiles := map[string]bool{
		"program_loading.go":              true,
		"update_actions_popup.go":         true,
		"update_pull_request_commands.go": true,
	}

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", regexp.MustCompile(`resetPullRequestListLoadState\(`), func(path string) bool {
		base := filepath.Base(path)
		return strings.HasSuffix(base, ".go") && !strings.HasSuffix(base, "_test.go")
	})

	remainingMatches := make([]string, 0, len(actualMatches))
	for _, match := range actualMatches {
		base := filepath.Base(strings.Split(match, ":")[0])
		if allowedFiles[base] {
			continue
		}
		remainingMatches = append(remainingMatches, match)
	}
	if len(remainingMatches) != 0 {
		t.Fatalf("expected pull request list load-state resets to stay behind the store helper and update-owned callers, actual %v", remainingMatches)
	}
}

func TestRefactorGuard_GivenProductionFiles_WhenScanning_ThenNoDirectDetailAndReviewChildStateFieldMutationRemains(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`program\.navigationState\.reviewSession\.selectedFileTreeRow\s*=\s*[^=]`,
		`program\.navigationState\.reviewSession\.collapsedTreeRowIDs\s*=\s*`,
		`program\.navigationState\.reviewSession\.collapsedThreadIDs\s*=\s*`,
		`program\.navigationState\.reviewSession\.collapsedTreeRowIDs\[[^\]]+\]\s*=`,
		`program\.navigationState\.reviewSession\.collapsedThreadIDs\[[^\]]+\]\s*=`,
		`program\.detailState\.wrapWidth\s*=\s*`,
		`program\.detailState\.viewState\.(?:cursor|preferredColumn)\s*=\s*`,
		`program\.detailState\.viewState\.(?:reset|sync)\(`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		base := filepath.Base(path)
		return strings.HasSuffix(base, ".go") && !strings.HasSuffix(base, "_test.go")
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected detail/review child-state field mutation to route through child reducers or whole-state replacement, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenFooterFile_WhenScanning_ThenOnlyViewGlueStillDependsOnProgram(t *testing.T) {
	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", regexp.MustCompile(`func \(program \*Program\)`), func(path string) bool {
		return filepath.Base(path) == "footer.go"
	})

	remainingMatches := make([]string, 0, len(actualMatches))
	for _, match := range actualMatches {
		if strings.Contains(match, "configurePaneFooterView(") || strings.Contains(match, "renderPaneFooterView(") {
			continue
		}
		remainingMatches = append(remainingMatches, match)
	}
	if len(remainingMatches) != 0 {
		t.Fatalf("expected footer helpers to stay on snapshot presenters instead of full Program coupling, actual %v", remainingMatches)
	}
}

func TestRefactorGuard_GivenHelpFile_WhenScanning_ThenOnlyViewGlueStillDependsOnProgram(t *testing.T) {
	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", regexp.MustCompile(`func \(program \*Program\)`), func(path string) bool {
		return filepath.Base(path) == "help.go"
	})

	remainingMatches := make([]string, 0, len(actualMatches))
	for _, match := range actualMatches {
		if strings.Contains(match, "configureHelpView(") || strings.Contains(match, "renderHelpView(") || strings.Contains(match, "fullPageHelpDown(") || strings.Contains(match, "fullPageHelpUp(") {
			continue
		}
		remainingMatches = append(remainingMatches, match)
	}
	if len(remainingMatches) != 0 {
		t.Fatalf("expected help overlay helpers to stay on snapshot presenters instead of full Program coupling, actual %v", remainingMatches)
	}
}

func TestRefactorGuard_GivenActionsPopupViewFile_WhenScanning_ThenOnlyViewGlueStillDependsOnProgram(t *testing.T) {
	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", regexp.MustCompile(`func \(program \*Program\)`), func(path string) bool {
		return filepath.Base(path) == "actions_popup_view.go"
	})

	remainingMatches := make([]string, 0, len(actualMatches))
	for _, match := range actualMatches {
		if strings.Contains(match, "layoutActionsPopupViews(") || strings.Contains(match, "layoutActionsPopupSearchView(") || strings.Contains(match, "configureActionsPopupChromeView(") || strings.Contains(match, "renderActionsPopupChromeView(") || strings.Contains(match, "configureActionsPopupView(") || strings.Contains(match, "renderActionsPopupView(") || strings.Contains(match, "configureActionsPopupSearchView(") || strings.Contains(match, "renderActionsPopupSearchView(") {
			continue
		}
		remainingMatches = append(remainingMatches, match)
	}
	if len(remainingMatches) != 0 {
		t.Fatalf("expected actions_popup_view.go to stay on snapshot presenters instead of full Program coupling, actual %v", remainingMatches)
	}
}

func TestRefactorGuard_GivenSearchViewFile_WhenScanning_ThenOnlyPromptGlueStillDependsOnProgram(t *testing.T) {
	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", regexp.MustCompile(`func \(program \*Program\)`), func(path string) bool {
		return filepath.Base(path) == "search_view.go"
	})

	remainingMatches := make([]string, 0, len(actualMatches))
	for _, match := range actualMatches {
		if strings.Contains(match, "layoutBottomPromptView(") || strings.Contains(match, "configureSearchView(") || strings.Contains(match, "renderSearchView(") || strings.Contains(match, "editSearch(") {
			continue
		}
		remainingMatches = append(remainingMatches, match)
	}
	if len(remainingMatches) != 0 {
		t.Fatalf("expected search_view.go to stay on prompt glue instead of full Program coupling, actual %v", remainingMatches)
	}
}

func TestRefactorGuard_GivenDetailAndReviewChildReducerFiles_WhenScanning_ThenTheyStayOnValueReceiverTransitions(t *testing.T) {
	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", regexp.MustCompile(strings.Join([]string{
		`func \(program \*Program\)`,
		`func \(state \*detailStateModel\)`,
		`func \(state \*reviewSessionState\)`,
	}, "|")), func(path string) bool {
		base := filepath.Base(path)
		return base == "detail_child_state.go" || base == "review_session_child_state.go"
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected detail/review child reducer files to stay on child-model value transitions, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenReviewSessionFiles_WhenScanning_ThenReadHelpersStayOnTheReadModel(t *testing.T) {
	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", regexp.MustCompile(`func \(program \*Program\)`), func(path string) bool {
		base := filepath.Base(path)
		return base == "review_session.go" || base == "review_session_content.go"
	})

	remainingMatches := make([]string, 0, len(actualMatches))
	for _, match := range actualMatches {
		if strings.Contains(match, "startReviewAction(") || strings.Contains(match, "executeStartReviewAction(") || strings.Contains(match, "exitReviewMode(") || strings.Contains(match, "reviewModePaneLayoutSize(") {
			continue
		}
		remainingMatches = append(remainingMatches, match)
	}
	if len(remainingMatches) != 0 {
		t.Fatalf("expected review-session read helpers to live on the focused read model instead of review_session*.go, actual %v", remainingMatches)
	}
}

func TestRefactorGuard_GivenUpdateInteractionFile_WhenScanning_ThenItDoesNotReachThroughGuiOrResolveViews(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`program\.gui`,
		`resolveView\(`,
		`currentDetailDocument\(`,
		`currentPullRequestBuildRunPopupDocument\(`,
		`syncDetailViewState\(`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		return filepath.Base(path) == "update_interaction.go"
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected update_interaction.go to stay on reducer intent selection without direct gui/view coupling, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenDetailSearchUpdateFiles_WhenScanning_ThenTheyDoNotReachThroughDetailViewShellHelpers(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`resolveView\(`,
		`currentDetailDocument\(`,
		`syncDetailViewState\(`,
		`followSubmittedDetailSearch\(`,
		`followReverseDetailSearch\(`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		base := filepath.Base(path)
		return base == "update_navigation_interaction.go" || base == "update.go"
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected detail-search update files to stop at typed shell commands instead of direct detail-view shell helpers, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenAsyncPopupAndRefreshUpdateFiles_WhenScanning_ThenTheyDoNotReachThroughGuiOrManualRefreshShellHelpers(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`program\.gui`,
		`program\.reportError\(`,
		`configureGUI\(`,
		`markManualPullRequestListRefresh\(`,
		`markManualPullRequestDetailRefresh\(`,
		`markManualPullRequestDiffRefresh\(`,
		`beginManualRefresh\(`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		base := filepath.Base(path)
		return base == "update_feature_async.go" || base == "update_actions_popup.go" || base == "update_async.go"
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected async popup/refresh update files to stop at typed shell commands instead of direct gui or manual-refresh shell helpers, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenRuntimeShortcutFiles_WhenScanning_ThenTheyUseShellHooksOrTypedCommandsInsteadOfGuiBranches(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`program\.gui`,
		`dispatch\(program\.gui`,
		`configureGUI\(`,
		`markManualNotificationRefresh\(`,
		`beginManualRefresh\(`,
		`reloadNotifications\(`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		base := filepath.Base(path)
		return base == "view_url.go" || base == "editor_dispatch.go" || base == "actions_popup_async_success.go" || base == "refresh_active_view.go"
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected runtime shortcut files to defer gui ownership and manual refresh work to shell hooks or explicit commands, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenUpdateReviewInteractionFile_WhenScanning_ThenItDoesNotReachThroughDetailViewMutationHelpers(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`resolveView\(`,
		`mutateDetailViewStateWithoutRefresh\(`,
		`toggleInlineConversationVisibilityState\(`,
		`setAllDetailFolds\(`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		return filepath.Base(path) == "update_review_interaction.go"
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected update_review_interaction.go to stop at typed shell commands instead of direct detail-view mutation, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenDetailFoldFiles_WhenScanning_ThenTheyDoNotReachThroughLiveDetailViews(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`resolveView\(`,
		`currentDetailDocument\(`,
		`syncDetailViewState\(`,
		`placeDetailCursorAtLine\(`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		base := filepath.Base(path)
		return base == "detail_bulk_fold.go" || base == "review_inline_conversation.go"
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected detail fold files to stay on reducer selectors while explicit commands own live detail-view work, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenCharacterMotionFiles_WhenScanning_ThenTheyDoNotReachThroughDetailShellHelpers(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`resolveView\(`,
		`currentDetailDocument\(`,
		`mutateDetailViewStateForYankMotion\(`,
		`mutatePullRequestBuildRunPopupViewStateForYankMotion\(`,
		`mutateDetailViewState\(`,
		`mutatePullRequestBuildRunPopupViewState\(`,
		`refreshShell\(`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		base := filepath.Base(path)
		return base == "program_character_motion.go" || base == "yank_motion.go"
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected character-motion and yank files to route live detail shell work through explicit commands, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenCursorDependentDetailActionFiles_WhenScanning_ThenTheyDoNotReachThroughLiveDetailDocuments(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`resolveView\(`,
		`currentDetailDocument\(`,
		`syncDetailViewState\(`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		base := filepath.Base(path)
		return base == "open_link.go" ||
			base == "pull_request_build.go" ||
			base == "pull_request_reviewer.go" ||
			base == "review_inline_comment.go" ||
			base == "reaction_target.go" ||
			base == "reaction_remove.go" ||
			base == "inline_comment_edit.go" ||
			base == "inline_comment_reply.go" ||
			base == "inline_comment_resolution.go"
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected cursor-dependent detail action files to reuse shared cursor selectors or commands instead of probing live detail documents directly, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenProgramNavigationFile_WhenScanning_ThenPageHandlersDispatchInsteadOfResolvingViews(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`resolveView\(`,
		`handlePageChange\(`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		return filepath.Base(path) == "program_navigation.go"
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected program_navigation.go page handlers to dispatch typed commands instead of resolving views inline, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenProgramNavigationFile_WhenScanning_ThenLineHandlersDispatchInsteadOfCallingSelectionShellHelpers(t *testing.T) {
	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", regexp.MustCompile(`handleSelectionChange\(`), func(path string) bool {
		return filepath.Base(path) == "program_navigation.go"
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected program_navigation.go line handlers to dispatch typed update paths instead of calling selection shell helpers, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenProgramNavigationFile_WhenScanning_ThenViewportHandlersDispatchInsteadOfOwningViewportShellHelpers(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`recenterListSelection\(`,
		`placeListSelection\(`,
		`scrollDown\(`,
		`scrollUp\(`,
		`recenter\(`,
		`placeCursorAtViewportTop\(`,
		`placeCursorAtViewportBottom\(`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		return filepath.Base(path) == "program_navigation.go"
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected program_navigation.go viewport handlers to dispatch typed commands instead of mutating viewport shell state inline, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenProgramNavigationSupportAndDetailSearchFiles_WhenScanning_ThenTheyStopOwningPageOrSearchShellHelpers(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`func \(program \*Program\) handlePageChange\(`,
		`mutateDetailViewStateForYankMotion\(`,
		`mutateDetailViewStateWithoutRefresh\(`,
		`followSubmittedDetailSearch\(`,
		`followReverseDetailSearch\(`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		base := filepath.Base(path)
		return base == "program_navigation_support.go" || base == "program_detail_search.go"
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected program_navigation_support.go and program_detail_search.go to leave page and detail-search shell helpers to explicit commands, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenCommandExecutorFiles_WhenScanning_ThenOnlyCmdExecuteAndBundleBuildersAcceptProgram(t *testing.T) {
	commandExecutorFiles := map[string]bool{
		"actions_popup_async_cmd.go":            true,
		"assignee_picker_search_cmd.go":         true,
		"cmd_actions_popup_async_requests.go":   true,
		"cmd_detail_fold.go":                    true,
		"cmd_detail_motion.go":                  true,
		"cmd_interaction.go":                    true,
		"cmd_modal_editor_submit_requests.go":   true,
		"cmd_popup_feature_request_requests.go": true,
		"cmd_popup_feature_requests.go":         true,
		"workflow_commands.go":                  true,
	}

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", regexp.MustCompile(`\*Program`), func(path string) bool {
		return commandExecutorFiles[filepath.Base(path)]
	})

	remainingMatches := make([]string, 0, len(actualMatches))
	for _, match := range actualMatches {
		if strings.Contains(match, "execute(program *Program,") || strings.Contains(match, "Deps(program *Program") || strings.Contains(match, "Runtime(program *Program") {
			continue
		}
		remainingMatches = append(remainingMatches, match)
	}
	if len(remainingMatches) != 0 {
		t.Fatalf("expected command executor files to depend on focused bundles instead of the full Program, actual %v", remainingMatches)
	}
}

func TestRefactorGuard_GivenWorkflowPlannerFile_WhenScanning_ThenItDoesNotDependOnProgramGuiOrInlineStoreMutation(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`\*Program`,
		`gocui\.Gui`,
		`program\.[A-Za-z0-9_]+`,
		`store\.[A-Za-z0-9_]+(?:\[[^\]]+\])?\s*=\s*`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		return filepath.Base(path) == "workflow_plans.go"
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected workflow_plans.go to stay on pure plan derivation without Program/GUI coupling or inline store mutation, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenProductionFiles_WhenScanning_ThenOnlyModelAndUpdateFilesUseProgramModelMutatorMethods(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`program\.model\.(?:Set|Open|Close|Select|Move|Update|Start|Cancel|Clear|Grow|Shrink|Focus|Blur|Submit|Advance|Cycle|Toggle|Reset|Remove|Add|Apply|Mark|Restore|Use)[A-Z][A-Za-z0-9_]*\(`,
		`program\.model\.adjustSelectionBy\(`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		base := filepath.Base(path)
		return strings.HasSuffix(base, ".go") && !strings.HasSuffix(base, "_test.go")
	})

	remainingMatches := make([]string, 0, len(actualMatches))
	for _, match := range actualMatches {
		base := filepath.Base(strings.Split(match, ":")[0])
		if strings.HasPrefix(base, "model") || strings.HasPrefix(base, "update") {
			continue
		}
		remainingMatches = append(remainingMatches, match)
	}
	if len(remainingMatches) != 0 {
		t.Fatalf("expected reducer/update files to own all program.model mutator methods, actual %v", remainingMatches)
	}
}

func TestRefactorGuard_GivenProductionFiles_WhenScanning_ThenOnlyShellGlueUsesDirectViewSyncHelpers(t *testing.T) {
	allowedFiles := map[string]bool{
		"program_loading.go":    true,
		"program_view_state.go": true,
		"shell_refresh.go":      true,
	}

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", regexp.MustCompile(`syncCurrentView\(|refreshDetailView\(|reloadActivePullRequestsTab\(`), func(path string) bool {
		base := filepath.Base(path)
		return strings.HasSuffix(base, ".go") && !strings.HasSuffix(base, "_test.go")
	})

	remainingMatches := make([]string, 0, len(actualMatches))
	for _, match := range actualMatches {
		base := filepath.Base(strings.Split(match, ":")[0])
		if allowedFiles[base] {
			continue
		}
		remainingMatches = append(remainingMatches, match)
	}
	if len(remainingMatches) != 0 {
		t.Fatalf("expected direct view-sync helpers to stay in shell glue, actual %v", remainingMatches)
	}
}

func TestRefactorGuard_GivenTUIPackageGoFiles_WhenScanning_ThenNoPhaseMigrationFileNamesRemain(t *testing.T) {
	packageRoot := given_guardPackageRoot(t)
	actualMatches := make([]string, 0)

	actualErr := filepath.WalkDir(packageRoot, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			name := entry.Name()
			if name == ".git" || name == "testdata" {
				return filepath.SkipDir
			}
			return nil
		}
		base := filepath.Base(path)
		if !strings.HasSuffix(base, ".go") || !strings.Contains(base, "_phase") {
			return nil
		}
		relativePath, relErr := filepath.Rel(packageRoot, path)
		if relErr != nil {
			return relErr
		}
		actualMatches = append(actualMatches, filepath.ToSlash(relativePath))
		return nil
	})
	if actualErr != nil {
		t.Fatalf("walk go files: %v", actualErr)
	}

	slices.Sort(actualMatches)
	if len(actualMatches) != 0 {
		t.Fatalf("expected migration phase file names to be removed, actual %v", actualMatches)
	}
}

func given_isRenderLayerFile(path string) bool {
	base := filepath.Base(path)
	if !strings.HasSuffix(base, ".go") || strings.HasSuffix(base, "_test.go") {
		return false
	}
	return strings.HasPrefix(base, "render") || strings.HasSuffix(base, "_view.go")
}

func given_regexpLineMatchesInGoFiles(t *testing.T, root string, pattern *regexp.Regexp, include func(string) bool) []string {
	t.Helper()

	packageRoot := given_guardPackageRoot(t)
	scanRoot := given_guardScanRoot(t, root)
	actualMatches := make([]string, 0)
	actualErr := filepath.WalkDir(scanRoot, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			name := entry.Name()
			if name == ".git" || name == "testdata" {
				return filepath.SkipDir
			}
			return nil
		}
		if include != nil && !include(path) {
			return nil
		}

		contents, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		relativePath, relErr := filepath.Rel(packageRoot, path)
		if relErr != nil {
			return relErr
		}
		for lineIndex, line := range strings.Split(string(contents), "\n") {
			if !pattern.MatchString(line) {
				continue
			}
			actualMatches = append(actualMatches, fmt.Sprintf("%s:%d contains %q", filepath.ToSlash(relativePath), lineIndex+1, strings.TrimSpace(line)))
		}
		return nil
	})
	if actualErr != nil {
		t.Fatalf("walk go files: %v", actualErr)
	}

	slices.Sort(actualMatches)
	return actualMatches
}
