package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"
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

func TestRefactorGuard_GivenSynchronousInteractionCommandFiles_WhenScanning_ThenTheyUseTheRuntimeBridgeInsteadOfNestedProgramDispatch(t *testing.T) {
	auditedFiles := map[string]bool{
		"cmd_interaction_modal.go":          true,
		"cmd_interaction_navigation.go":     true,
		"cmd_interaction_detail_search.go":  true,
		"cmd_detail_fold.go":                true,
		"cmd_detail_motion.go":              true,
		"cmd_interaction_link_clipboard.go": true,
		"cmd_interaction_io.go":             true,
	}

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", regexp.MustCompile(`program\.dispatch`), func(path string) bool {
		return auditedFiles[filepath.Base(path)]
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected synchronous interaction command re-entry to use executeRuntimeMessage(...) instead of nested program.dispatch(...), actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenUpdateFiles_WhenScanning_ThenTheyDoNotUseTheGlobalActionsPopupResyncDefer(t *testing.T) {
	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", regexp.MustCompile(strings.Join([]string{
		`defer program\.resyncVisibleActionsPopupSearchInUpdate\(`,
		`func \(program \*Program\) resyncVisibleActionsPopupSearchInUpdate\(`,
	}, "|")), func(path string) bool {
		base := filepath.Base(path)
		return base == "update.go" || base == "update_actions_popup.go"
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected popup-search filtered-index maintenance to stay on explicit reducer/read-side seams instead of the global update defer, actual %v", actualMatches)
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
		if strings.HasPrefix(base, "update") || strings.HasPrefix(base, "cmd_interaction") {
			continue
		}
		remainingMatches = append(remainingMatches, match)
	}
	if len(remainingMatches) != 0 {
		t.Fatalf("expected direct feedback and error reporting to stay in update files or explicit interaction commands, actual %v", remainingMatches)
	}
}

func TestRefactorGuard_GivenProductionFiles_WhenScanning_ThenNoLegacyReportErrorCommandOrShellMutationHelperRemains(t *testing.T) {
	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", regexp.MustCompile(`reportErrorCmd|program\.reportError\(`), func(path string) bool {
		base := filepath.Base(path)
		return strings.HasSuffix(base, ".go") && !strings.HasSuffix(base, "_test.go")
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected transient error popup reporting to flow through typed update messages instead of reportErrorCmd or program.reportError(...), actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenProductionFiles_WhenScanning_ThenNoDeadShellRefreshOrLoadingHelpersRemain(t *testing.T) {
	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", regexp.MustCompile(strings.Join([]string{
		`refreshShell\(`,
		`refreshDetailView\(`,
		`mutateDetailViewState\(`,
		`mutateDetailViewStateWithoutRefresh\(`,
		`maybeLoadConnectedUser\(`,
		`maybeLoadActivePullRequests\(`,
		`maybeLoadPullRequests\(`,
		`reloadActivePullRequestsTab\(`,
		`reloadNotifications\(`,
	}, "|")), func(path string) bool {
		base := filepath.Base(path)
		return strings.HasSuffix(base, ".go") && !strings.HasSuffix(base, "_test.go")
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected dead shell refresh and loading helper leftovers to be removed from production code, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenProductionFiles_WhenScanning_ThenNoDeadMaybeLoadWorkflowWrappersRemain(t *testing.T) {
	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", regexp.MustCompile(strings.Join([]string{
		`func \(program \*Program\) maybeLoadSelectedPullRequestDetail\(`,
		`func \(program \*Program\) maybeLoadSelectedPullRequestDiff\(`,
		`func \(program \*Program\) maybeLoadNotifications\(`,
		`func \(program \*Program\) maybeLoadSelectedNotificationDetail\(`,
		`func \(program \*Program\) maybeLoadCurrentDetailImageHTML\(`,
		`func \(program \*Program\) maybeLoadCurrentDetailImages\(`,
	}, "|")), func(path string) bool {
		base := filepath.Base(path)
		return strings.HasSuffix(base, ".go") && !strings.HasSuffix(base, "_test.go")
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected dead maybeLoad workflow wrappers to be removed from production code, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenProductionFiles_WhenScanning_ThenNoDeadRepeatSearchHelpersRemain(t *testing.T) {
	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", regexp.MustCompile(strings.Join([]string{
		`func \(program \*Program\) repeatActionsPopupSearch\(`,
		`func \(program \*Program\) repeatSideSearch\(`,
		`func \(program \*Program\) repeatReviewFileTreeSearch\(`,
	}, "|")), func(path string) bool {
		base := filepath.Base(path)
		return strings.HasSuffix(base, ".go") && !strings.HasSuffix(base, "_test.go")
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected dead repeat-search helpers to be removed from production code, actual %v", actualMatches)
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

func TestRefactorGuard_GivenProductionFiles_WhenScanning_ThenOnlyActionsPopupAsyncPreflightHelpersInstantiateTheAsyncCommand(t *testing.T) {
	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", regexp.MustCompile(`actionsPopupAsyncCmd\s*\{\s*request\s*:`), func(path string) bool {
		base := filepath.Base(path)
		return strings.HasSuffix(base, ".go") && !strings.HasSuffix(base, "_test.go")
	})

	remainingMatches := make([]string, 0, len(actualMatches))
	for _, match := range actualMatches {
		base := filepath.Base(strings.Split(match, ":")[0])
		if base == "update_actions_popup_async.go" {
			continue
		}
		remainingMatches = append(remainingMatches, match)
	}
	if len(remainingMatches) != 0 {
		t.Fatalf("expected actions-popup async commands to be queued only through the reducer-owned preflight helper, actual %v", remainingMatches)
	}
}

func TestRefactorGuard_GivenActionsPopupAsyncCommandFile_WhenScanning_ThenItAvoidsReducerOwnedPopupPreflightStateMutation(t *testing.T) {
	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", regexp.MustCompile(strings.Join([]string{
		`actionsPopupWidget\.errorMessage\s*=`,
		`startGHCommandLoading\(`,
	}, "|")), func(path string) bool {
		return filepath.Base(path) == "actions_popup_async_cmd.go"
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected actions_popup_async_cmd.go to stay on IO plus finish dispatch, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenAsyncCompletionFiles_WhenScanning_ThenTheyDoNotEmbedSuccessMessagesOrNestedReducerRecursion(t *testing.T) {
	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", regexp.MustCompile(strings.Join([]string{
		`Success\s+Msg`,
		`Success:`,
		`Update\(program,\s*message\.Success\)`,
	}, "|")), func(path string) bool {
		base := filepath.Base(path)
		return base == "msg_feature_async.go" || base == "actions_popup_async_cmd.go" || base == "update_feature_async.go"
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected async completion wrappers to carry typed completion data instead of nested success messages or recursive Update calls, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenModalCompletionFiles_WhenScanning_ThenTheyDoNotEmbedSuccessMessagesOrNestedReducerRecursion(t *testing.T) {
	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", regexp.MustCompile(strings.Join([]string{
		`Success\s+Msg`,
		`Success:`,
		`Update\(program,\s*message\.Success\)`,
	}, "|")), func(path string) bool {
		base := filepath.Base(path)
		return base == "msg_interaction.go" || base == "cmd_interaction_modal.go" || base == "update_interaction.go"
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected modal submit completion wrappers to carry typed completion data instead of nested success messages or recursive Update calls, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenProgramViewStateFile_WhenScanning_ThenAfterStateChangeStaysShellOnly(t *testing.T) {
	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", regexp.MustCompile(strings.Join([]string{
		`clearExpiredYankHighlights\(`,
		`clearExpiredTransientErrorPopup\(`,
		`syncActionsPopupSearch\(`,
	}, "|")), func(path string) bool {
		return filepath.Base(path) == "program_view_state.go"
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected afterStateChange to stay on workflow planning, shell sync, and redraw only, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenUpdateHelperFiles_WhenScanning_ThenTheyDoNotReenterUpdateForLocalFollowUp(t *testing.T) {
	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", regexp.MustCompile(`\bUpdate\(program,`), func(path string) bool {
		base := filepath.Base(path)
		return strings.HasPrefix(base, "update") && strings.HasSuffix(base, ".go") && !strings.HasSuffix(base, "_test.go") && base != "update.go"
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected update helper files to stop re-entering Update(program, ...) for local follow-up behavior, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenUpdateRouter_WhenScanning_ThenTopLevelRoutingStaysGroupedByNamedCategories(t *testing.T) {
	contents, actualErr := os.ReadFile("update.go")
	then_noError(t, actualErr)
	actualSource := string(contents)

	expectedOrder := []string{
		`if result := program.routeLifecycleAndEditorMessages(msg); result.handled {`,
		`if result := program.routeNavigationAndDetailMessages(msg); result.handled {`,
		`if result := program.routeBrowserClipboardAndLinkMessages(msg); result.handled {`,
		`if result := program.routeNotificationAndSearchMessages(msg); result.handled {`,
		`if result := program.routeMutationMessages(msg); result.handled {`,
		`if result := program.routeWorkflowMessages(msg); result.handled {`,
		`if result := program.routeActionsPopupMessages(msg); result.handled {`,
	}

	missing := make([]string, 0)
	lastIndex := -1
	for _, fragment := range expectedOrder {
		actualIndex := strings.Index(actualSource, fragment)
		if actualIndex < 0 {
			missing = append(missing, fragment)
			continue
		}
		if actualIndex <= lastIndex {
			t.Fatalf("expected update.go route categories to stay in order %v", expectedOrder)
		}
		lastIndex = actualIndex
	}
	if len(missing) != 0 {
		t.Fatalf("expected update.go to route through named message categories, missing %v", missing)
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

func TestRefactorGuard_GivenModalEditorFiles_WhenScanning_ThenTheyUseTypedSubmitDescriptorsInsteadOfFunctionFields(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`submitRequested\s+func\(string\)\s+Msg`,
		`WithSubmitRequested\(`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		base := filepath.Base(path)
		return strings.HasSuffix(base, ".go") && !strings.HasSuffix(base, "_test.go")
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected modal editor state and open helpers to use typed submit descriptors instead of function fields, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenMsgAndCommandFiles_WhenScanning_ThenTheyDoNotExposeLiveViewPointersOnSurfaceTypes(t *testing.T) {
	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", regexp.MustCompile(`^\s*View\s+\*gocui\.View`), func(path string) bool {
		base := filepath.Base(path)
		return strings.HasSuffix(base, ".go") && !strings.HasSuffix(base, "_test.go") && (strings.HasPrefix(base, "msg") || strings.HasPrefix(base, "cmd"))
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected msg and cmd surfaces to keep live gocui views out of their durable fields, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenAuditedReadHelpers_WhenScanning_ThenTheyDoNotKeepStaleLiveViewParameters(t *testing.T) {
	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", regexp.MustCompile(strings.Join([]string{
		`reviewSessionCommentTarget\(detailView \*gocui\.View`,
		`reviewSessionCommentLocations\(detailView \*gocui\.View`,
		`currentDetailCursorLink\(_ \*gocui\.View`,
	}, "|")), func(path string) bool {
		base := filepath.Base(path)
		return base == "review_navigation.go" || base == "open_link.go"
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected the audited read helpers to stop advertising stale live-view parameters, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenEditorStateFiles_WhenScanning_ThenTheyDoNotStorePointerBackedEditorsOrModalPointerPayloads(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`modalEditor\s+\*modalEditorState`,
		`editor\s+\*lineEditor`,
		`searchEditor\s+\*lineEditor`,
		`lineEditor\s+\*lineEditor`,
		`editor\s+\*multilineEditor`,
		`State\s+\*modalEditorState`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		base := filepath.Base(path)
		return base == "program_state_models.go" || base == "widget_state.go" || base == "modal_editor.go" || base == "msg_interaction.go"
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected overlay, widget, and modal-open state to use value models instead of pointer-backed editor fields, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenCmdInteractionFile_WhenScanning_ThenItDoesNotOwnTheGiantRuntimeBundle(t *testing.T) {
	contents, actualErr := os.ReadFile("cmd_interaction.go")
	then_noError(t, actualErr)
	actualSource := string(contents)

	for _, forbiddenSnippet := range []string{
		"type interactionCommandRuntime struct",
		"func newInteractionCommandRuntime(program *Program) interactionCommandRuntime",
	} {
		if strings.Contains(actualSource, forbiddenSnippet) {
			t.Fatalf("expected cmd_interaction.go to shed the old multi-domain runtime bundle, actual source:\n%s", actualSource)
		}
	}
}

func TestRefactorGuard_GivenProductionFiles_WhenScanning_ThenNoPopupToModalEditorCallbackBridgeRemains(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`openModalEditorFromActionsPopup\(`,
		`open\s+func\(\*gocui\.Gui\)\s+error`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		base := filepath.Base(path)
		return strings.HasSuffix(base, ".go") && !strings.HasSuffix(base, "_test.go")
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected popup-triggered modal editor opens to use typed open requests instead of callback bridges, actual %v", actualMatches)
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

func TestRefactorGuard_GivenDetailImageHTMLFiles_WhenScanning_ThenTheyUseTypedApplyTargetsInsteadOfCallbacks(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`applyRenderedHTML\s+func\(\*Program,\s*string\)`,
		`Source\.applyRenderedHTML\(`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		base := filepath.Base(path)
		return base == "detail_image_loader.go" || base == "update_async.go"
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected detail-image HTML load results to use typed apply targets instead of callbacks, actual %v", actualMatches)
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
		"workflow_session_commands.go":             true,
		"workflow_pull_request_list_commands.go":   true,
		"workflow_pull_request_detail_commands.go": true,
		"workflow_notification_commands.go":        true,
		"workflow_detail_image_commands.go":        true,
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

func TestRefactorGuard_GivenProductionFiles_WhenScanning_ThenWorkflowCommandRuntimesAvoidTheSharedCrossDomainDepsBag(t *testing.T) {
	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", regexp.MustCompile(`type workflowCommandDeps struct|newWorkflowCommandDeps\(`), func(path string) bool {
		base := filepath.Base(path)
		return strings.HasSuffix(base, ".go") && !strings.HasSuffix(base, "_test.go")
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected workflow command runtimes to use domain-scoped bundles instead of the shared workflowCommandDeps bag, actual %v", actualMatches)
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

func TestRefactorGuard_GivenProductionFiles_WhenScanning_ThenOpenedPullRequestPinningStaysInUpdateOwnedHelpers(t *testing.T) {
	allowedFiles := map[string]bool{
		"update_pull_request_navigation_state.go": true,
	}

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", regexp.MustCompile(strings.Join([]string{
		`program\.navigationState\.openedPullRequestSummary\s*=\s*[^=]`,
		`program\.navigationState\.openedPullRequestTab\s*=\s*[^=]`,
	}, "|")), func(path string) bool {
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
		t.Fatalf("expected opened pull request summary pinning to stay in update-owned navigation helpers, actual %v", remainingMatches)
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

func TestRefactorGuard_GivenFooterPresenterAdapter_WhenScanning_ThenItAvoidsResolvingFullActionsPopupLists(t *testing.T) {
	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", regexp.MustCompile(`currentActionsPopupActions\(`), func(path string) bool {
		return filepath.Base(path) == "footer_presenter_adapter.go"
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected footer_presenter_adapter.go to stay on cheap action availability hints instead of resolving the full popup action list, actual %v", actualMatches)
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

func TestRefactorGuard_GivenStatusLineFiles_WhenScanning_ThenTheyUseSnapshotPresenterValuesInsteadOfProgramBackedStatusSelection(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`type StatusLinePresenter struct`,
		`func \(program \*Program\) statusLineText\(`,
		`func \(program \*Program\) loadingStatusText\(`,
		`presenter\.program\.statusLineText\(`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		base := filepath.Base(path)
		return base == "status_line.go" || base == "render_pipeline.go"
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected status-line text selection to flow through snapshot presenter values instead of full Program coupling, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenHelpFile_WhenScanning_ThenHelpPagingDispatchesTypedRequests(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`resolveView\(`,
		`scrollReadOnlyView\(`,
		`viewPageSize\(`,
		`executeCmds\(`,
		`readOnlyScrollCmd\s*\{`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		return filepath.Base(path) == "help.go"
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected help.go paging handlers to dispatch typed help-scroll requests instead of executing read-only scroll commands inline, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenRenderPipelineFile_WhenScanning_ThenLayoutPlanningUsesSnapshotInputsInsteadOfProgramBackedPlannerStructs(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`type MainPanelRenderer struct`,
		`type SidePanelRenderer struct`,
		`type OverlayRenderer struct`,
		`type KeyHintPresenter struct`,
		`func \(program \*Program\) mainPanelRenderer\(`,
		`func \(program \*Program\) sidePanelRenderer\(`,
		`func \(program \*Program\) overlayRenderer\(`,
		`func \(program \*Program\) keyHintPresenter\(`,
		`func \(program \*Program\) screenLayoutForSize\(`,
		`func \(program \*Program\) screenCompositionForSize\(`,
		`func \(program \*Program\) paneFooterTextForView\(`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		return filepath.Base(path) == "render_pipeline.go"
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected render_pipeline.go to plan layout and composition from snapshot inputs instead of Program-backed planner structs, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenProgramNavigationSupportFile_WhenScanning_ThenItDoesNotOwnReadOnlyScrollHelpers(t *testing.T) {
	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", regexp.MustCompile(`scrollReadOnlyView\(`), func(path string) bool {
		return filepath.Base(path) == "program_navigation_support.go"
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected program_navigation_support.go to stop owning help-specific read-only scroll helpers, actual %v", actualMatches)
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

func TestRefactorGuard_GivenAssigneePickerFiles_WhenScanning_ThenNestedStateWritesStayOnTheChildStateAdapter(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`program\.actionsPopupWidget\.assigneePicker\.[A-Za-z0-9_]+\s*=\s*[^=]`,
		`delete\(program\.actionsPopupWidget\.assigneePicker\.selectedLogins`,
		`program\.actionsPopupWidget\.assigneePicker\.rememberCandidates\(`,
		`program\.actionsPopupWidget\.assigneePicker\s*=\s*newAssigneePickerState\(`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		base := filepath.Base(path)
		return base == "update_actions_popup.go" || base == "update_feature_async.go" || base == "pull_request_assignee.go"
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected assignee-picker state writes to stay on a dedicated child-state adapter instead of nested reducer mutation, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenManualRefreshFiles_WhenScanning_ThenNestedStateWritesStayOnValueTransitions(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`program\.manualRefreshState\.(?:pullRequestListPending|pullRequestDetailPending|pullRequestDiffPending|notificationPending|feedback)\s*=\s*[^=]`,
		`program\.manualRefreshState\.(?:pullRequestListPending|pullRequestDetailPending|pullRequestDiffPending)\[[^\]]+\]\s*=`,
		`delete\(program\.manualRefreshState\.(?:pullRequestListPending|pullRequestDetailPending|pullRequestDiffPending)`,
		`program\.feedbackMessage\s*=\s*[^=]`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		base := filepath.Base(path)
		return base == "manual_refresh_notifications.go" || base == "manual_refresh_preflight.go" || base == "pull_request_refresh_reporting.go"
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected manual-refresh bookkeeping to use value transitions plus whole-state replacement instead of nested state writes, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenActionsPopupChromeFiles_WhenScanning_ThenNestedStateWritesStayOnWholeStateReplacementHelpers(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`program\.actionsPopupWidget\.(?:errorMessage|pendingConfirmationActionID|reactionPicker|themePicker|assigneePicker|assigneePickerLoad)\s*=\s*[^=]`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		base := filepath.Base(path)
		return strings.HasSuffix(base, ".go") && !strings.HasSuffix(base, "_test.go")
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected actions-popup chrome state writes to use whole-state replacement helpers instead of nested field mutation, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenStatusStoreFiles_WhenScanning_ThenSharedStatusFieldsStayOnValueTransitions(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`program\.(?:feedbackMessage|ghCommandLoadingMessage|storyReviewLoading)\s*=\s*[^=]`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		base := filepath.Base(path)
		return strings.HasSuffix(base, ".go") && !strings.HasSuffix(base, "_test.go")
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected shared status-store writes to use value transitions plus whole-store replacement instead of direct field mutation, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenSessionStoreFiles_WhenScanning_ThenConnectedUserWritesStayOnValueTransitions(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`program\.(?:connectedUserLoadStarted|connectedUserLogin|connectedUserName)\s*=\s*[^=]`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		base := filepath.Base(path)
		return strings.HasSuffix(base, ".go") && !strings.HasSuffix(base, "_test.go")
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected connected-user session-store writes to use value transitions plus whole-store replacement instead of direct field mutation, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenSearchWidgetFiles_WhenScanning_ThenSearchEditorAndDirectionWritesStayOnWholeStateReplacementHelpers(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`program\.searchWidget\.(?:openEditor|clearEditor)\(`,
		`program\.searchWidget\.editor\.ApplyIntent\(`,
		`program\.searchWidget\.detailReversed\s*=\s*[^=]`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		base := filepath.Base(path)
		return strings.HasSuffix(base, ".go") && !strings.HasSuffix(base, "_test.go")
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected search-widget editor and direction writes to use whole-state replacement helpers instead of mutable helper calls or direct field mutation, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenActionsPopupWidgetFiles_WhenScanning_ThenPopupSearchWritesStayOnWholeStateReplacementHelpers(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`program\.actionsPopupWidget\.(?:openSearchEditor|clearSearchEditor)\(`,
		`program\.actionsPopupWidget\.searchEditor\.ApplyIntent\(`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		base := filepath.Base(path)
		return strings.HasSuffix(base, ".go") && !strings.HasSuffix(base, "_test.go")
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected popup-search editor writes to use whole-state replacement helpers instead of mutable helper calls or direct field mutation, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenDetailStateFiles_WhenScanning_ThenActiveDetailTabWritesStayOnWholeStateReplacementHelpers(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(`program\.detailState\.activeTab\s*=\s*[^=]`)

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		base := filepath.Base(path)
		return strings.HasSuffix(base, ".go") && !strings.HasSuffix(base, "_test.go")
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected active detail-tab writes to use whole-state replacement helpers instead of direct field mutation, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenDetailStateFiles_WhenScanning_ThenDetailPrefixAndVisualCleanupStayOnWholeStateReplacementHelpers(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`program\.detailState\.viewState\.(?:clearPendingPrefix|exitVisualMode)\(`,
		`detailState\.viewState\.(?:clearPendingPrefix|exitVisualMode)\(`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		base := filepath.Base(path)
		return strings.HasSuffix(base, ".go") && !strings.HasSuffix(base, "_test.go")
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected detail pending-prefix and visual-mode cleanup to use whole-state replacement helpers instead of nested detail-view mutation, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenReviewSessionFiles_WhenScanning_ThenLifecycleSelectionAndSummaryWritesStayOnChildStateAdapters(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`program\.navigationState\.reviewSession\s*=\s*[^=]`,
		`mutate\(&program\.navigationState\.reviewSession\.summary\)`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		base := filepath.Base(path)
		return strings.HasSuffix(base, ".go") && !strings.HasSuffix(base, "_test.go") && base != "review_session_child_state_adapter.go"
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected review-session lifecycle, selection, and summary writes to stay on the child-state adapter surface instead of update files or nested summary mutation, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenRuntimeConfigFiles_WhenScanning_ThenRuntimeConfigWritesStayOnWholeStateReplacementHelpers(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`program\.runtimeConfig\.(?:pullRequestSearches|keymapOverrides|storyReviewConfig)\s*=\s*[^=]`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		base := filepath.Base(path)
		return strings.HasSuffix(base, ".go") && !strings.HasSuffix(base, "_test.go")
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected runtime-config writes to use whole-state replacement helpers instead of direct field mutation, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenNotificationStoreFiles_WhenScanning_ThenNotificationLoadingWritesStayOnValueTransitions(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`program\.(?:notificationsLoadStarted|notificationsLoading|notificationsLoadingDetailMessage)\s*=\s*[^=]`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		base := filepath.Base(path)
		return strings.HasSuffix(base, ".go") && !strings.HasSuffix(base, "_test.go")
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected notification loading writes to use value transitions plus whole-store replacement instead of direct field mutation, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenBuildStoreFiles_WhenScanning_ThenBuildLoadAndPopupWritesStayOnValueTransitions(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`program\.(?:pullRequestBuildRunLoad|pullRequestBuildRunPopup)\s*=\s*[^=]`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		base := filepath.Base(path)
		return strings.HasSuffix(base, ".go") && !strings.HasSuffix(base, "_test.go")
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected build-store load and popup writes to use value transitions plus whole-store replacement instead of direct field mutation, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenModalEditorFiles_WhenScanning_ThenModalEditorErrorWritesStayOnWholeStateReplacementHelpers(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`program\.overlayState\.modalEditor\.errorMessage\s*=\s*[^=]`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		base := filepath.Base(path)
		return strings.HasSuffix(base, ".go") && !strings.HasSuffix(base, "_test.go")
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected modal-editor error writes to use whole-state replacement helpers instead of direct child-state mutation, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenReviewSessionFiles_WhenScanning_ThenReadHelpersStayOnTheReadModel(t *testing.T) {
	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", regexp.MustCompile(`func \(program \*Program\)`), func(path string) bool {
		base := filepath.Base(path)
		return base == "review_session.go" || base == "review_session_content.go"
	})

	remainingMatches := make([]string, 0, len(actualMatches))
	for _, match := range actualMatches {
		if strings.Contains(match, "startReviewAction(") || strings.Contains(match, "exitReviewMode(") || strings.Contains(match, "reviewModePaneLayoutSize(") {
			continue
		}
		remainingMatches = append(remainingMatches, match)
	}
	if len(remainingMatches) != 0 {
		t.Fatalf("expected review-session read helpers to live on the focused read model instead of review_session*.go, actual %v", remainingMatches)
	}
}

func TestRefactorGuard_GivenReviewSessionReadModelAdapter_WhenScanning_ThenItAvoidsEagerReviewDescriptionOverviewRendering(t *testing.T) {
	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", regexp.MustCompile(`renderCurrentPullRequestOverview\(`), func(path string) bool {
		return filepath.Base(path) == "review_session_read_model_adapter.go"
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected review_session_read_model_adapter.go to keep overview rendering on lazy read-model helpers instead of eager adapter construction, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenReviewSessionReadModelAdapter_WhenScanning_ThenItDoesNotMutateDurableReviewSelectionState(t *testing.T) {
	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", regexp.MustCompile(`program\.navigationState\.reviewSession\s*=\s*`), func(path string) bool {
		return filepath.Base(path) == "review_session_read_model_adapter.go"
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected review_session_read_model_adapter.go to stay read-only and leave durable review selection mutation to child-state helpers, actual %v", actualMatches)
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

func TestRefactorGuard_GivenRefreshCommandFile_WhenScanning_ThenItAvoidsReducerOwnedManualRefreshPreflightMutation(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`markManualPullRequestListRefresh\(`,
		`markManualPullRequestDetailRefresh\(`,
		`markManualPullRequestDiffRefresh\(`,
		`markManualNotificationRefresh\(`,
		`beginManualRefresh\(`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		return filepath.Base(path) == "cmd_interaction_refresh.go"
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected cmd_interaction_refresh.go to stay on reload execution instead of reducer-owned manual refresh preflight mutation, actual %v", actualMatches)
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

func TestRefactorGuard_GivenDetailFoldHelperFiles_WhenScanning_ThenTheyDoNotClearPendingPrefixesDirectly(t *testing.T) {
	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", regexp.MustCompile(`detailState\.viewState\.clearPendingPrefix\(`), func(path string) bool {
		base := filepath.Base(path)
		return base == "detail_bulk_fold.go" || base == "review_inline_conversation.go"
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected detail-fold helper files to leave pending-prefix teardown to update-owned handlers, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenDetailFoldCommandFile_WhenScanning_ThenItDoesNotMutateDurableDetailStateOrWireReducerOwnedCollapseHelpers(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`syncDetailViewState\(`,
		`placeDetailCursorAtLine\(`,
		`applyDetailViewSyncPlan\(`,
		`toggleInlineConversationVisibilityState\(`,
		`setAllDetailFolds\(`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		return filepath.Base(path) == "cmd_detail_fold.go"
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected cmd_detail_fold.go to stay on live detail lookup plus typed resolved-message dispatch instead of reducer-owned collapse helpers, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenLinkClipboardCommandFile_WhenScanning_ThenItDoesNotMutateDetailOrBuildPopupStateOrDispatchFeedbackDirectly(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`syncDetailViewState\(`,
		`program\.detailState\.viewState\.`,
		`popup\.viewState\.`,
		`syncPullRequestBuildRunPopupViewState\(`,
		`MsgFeedbackSet\s*\{`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		return filepath.Base(path) == "cmd_interaction_link_clipboard.go"
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected cmd_interaction_link_clipboard.go to resolve live documents and links only while reducer-owned messages handle feedback, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenAuditedShellBridgeFiles_WhenScanning_ThenOnlyRuntimeDispatchKeepsDirectUpdateCalls(t *testing.T) {
	auditedFiles := map[string]bool{
		"actions_popup_async_cmd.go":    true,
		"cmd_popup_feature_requests.go": true,
		"workflow_command_runtime.go":   true,
		"pull_request_commands.go":      true,
		"runtime_dispatch.go":           true,
	}
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`Update\(program, Msg`,
		`executeCmds\([^\n]*Update\(program, msg\)`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		return auditedFiles[filepath.Base(path)]
	})

	remainingMatches := make([]string, 0, len(actualMatches))
	for _, match := range actualMatches {
		base := filepath.Base(strings.Split(match, ":")[0])
		if base == "runtime_dispatch.go" {
			continue
		}
		remainingMatches = append(remainingMatches, match)
	}
	if len(remainingMatches) != 0 {
		t.Fatalf("expected audited shell-side message hops to route through the runtime dispatch bridge instead of inline Update(...) calls, actual %v", remainingMatches)
	}
}

func TestRefactorGuard_GivenModalEditorSubmitRequestFile_WhenScanning_ThenItDoesNotWirePendingReviewStoreMutation(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`recordPendingPullRequestReview`,
		`setPendingPullRequestReviewStateByIdentity`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		return filepath.Base(path) == "cmd_modal_editor_submit_requests.go"
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected cmd_modal_editor_submit_requests.go to stop wiring pending-review store mutation directly, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenReviewCollapseAndBrowserSectionFiles_WhenScanning_ThenStateWritesStayOnReducerOwnedHelpers(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`program\.navigationState\.reviewSession\s*=`,
		`program\.browserCollapsedSectionStates\s*=`,
		`program\.browserCollapsedSectionStates\[[^\]]+\]\s*=`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		base := filepath.Base(path)
		return base == "review_inline_conversation.go" || base == "detail_bulk_fold.go" || base == "review_file_tree_search.go" || base == "browser_detail_sections.go"
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected review collapse and browser section feature files to stop mutating reducer-owned state directly, actual %v", actualMatches)
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

func TestRefactorGuard_GivenDetailCharacterMotionBindingsFile_WhenScanning_ThenItUsesSelectorSnapshotsInsteadOfLiveViews(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`program\.gui`,
		`resolveView\(`,
		`viewPageSize\(`,
		`syncDetailViewState\(`,
		`syncPullRequestBuildRunPopupViewState\(`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		return filepath.Base(path) == "detail_character_motion_bindings.go"
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected detail_character_motion_bindings.go to derive target bindings from selector snapshots instead of live views, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenBuildPopupNavigationFiles_WhenScanning_ThenTheyUseExplicitPopupMotionCommands(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`resolveView\(`,
		`currentPullRequestBuildRunPopupDocument\(`,
		`mutatePullRequestBuildRunPopupViewStateWithoutRefresh\(`,
		`mutatePullRequestBuildRunPopupViewStateForYankMotion\(`,
		`mutatePullRequestBuildRunPopupViewState\(`,
		`refreshShell\(`,
		`followSubmittedPullRequestBuildRunPopupSearch\(`,
		`repeatPullRequestBuildRunPopupSearch\(`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		base := filepath.Base(path)
		return base == "pull_request_build_popup_navigation.go" || base == "pull_request_build_popup_search.go" || base == "pull_request_build_popup_search_repeat.go"
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected build-popup navigation and search files to defer live popup shell work to explicit commands, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenPullRequestBuildPopupFile_WhenScanning_ThenRenderAndLinkHelpersUseRenderPrepOrSnapshotSelectors(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`viewState\.syncSearch\(`,
		`syncPullRequestBuildRunPopupViewState\(`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		return filepath.Base(path) == "pull_request_build_popup.go"
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected pull_request_build_popup.go render and link helpers to stay on render prep or snapshot selectors instead of inline popup state sync, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenDetailRenderPrepFiles_WhenScanning_ThenPrepareViewRenderStateStaysReadOnly(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`syncDetailViewShellState\(`,
		`syncPullRequestBuildRunPopupShellState\(`,
		`program\.detailState\s*=\s*[^=]`,
		`pullRequestBuildRunPopup\.viewState\.sync\(`,
		`pullRequestBuildRunPopup\.viewState\.syncSearch\(`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		return filepath.Base(path) == "detail_render_state.go"
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected detail_render_state.go to stay read-only while shell sync owns durable detail/build-popup clamping, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenRenderRefreshFiles_WhenScanning_ThenTheyReuseSyncViewShellStateInsteadOfDirectRenderStateSync(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`syncDetailViewShellState\(`,
		`syncPullRequestBuildRunPopupShellState\(`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		base := filepath.Base(path)
		return base == "render_pipeline.go" || base == "program_view_state.go"
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected render refresh files to funnel detail/build-popup shell sync through syncViewShellState(...), actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenActionsPopupInteractionFile_WhenScanning_ThenItStopsOwningPopupPageAndViewportShellHelpers(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`resolveView\(`,
		`recenterListSelection\(`,
		`placeListSelection\(`,
		`viewPageSize\(`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		return filepath.Base(path) == "actions_popup_interaction.go"
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected actions_popup_interaction.go to stop at popup-selection intent while explicit commands own page and viewport shell helpers, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenActionsPopupFiles_WhenScanning_ThenTheyUseTypedRequestsInsteadOfExecuteCallbacks(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`execute\s+func\(\*gocui\.Gui\) error`,
		`actionsPopupExecuteErr\(`,
		`action\.execute\(gui\)`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		base := filepath.Base(path)
		return base == "actions_popup.go" || base == "actions_popup_interaction.go"
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected actions popup surfaces to carry typed requests instead of execute callbacks, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenProductionFiles_WhenScanning_ThenOnlyTheActionsPopupSubmitEntrypointKeepsAnExecuteActionName(t *testing.T) {
	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", regexp.MustCompile(`func \(program \*Program\) execute[A-Za-z0-9]+Action\(`), func(path string) bool {
		base := filepath.Base(path)
		return strings.HasSuffix(base, ".go") && !strings.HasSuffix(base, "_test.go")
	})

	remainingMatches := make([]string, 0, len(actualMatches))
	for _, match := range actualMatches {
		if strings.Contains(match, "executeSelectedActionsPopupAction(") {
			continue
		}
		remainingMatches = append(remainingMatches, match)
	}
	if len(remainingMatches) != 0 {
		t.Fatalf("expected popup-only execute*Action helpers to be deleted or renamed, actual %v", remainingMatches)
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

func TestRefactorGuard_GivenNavigationCommandFile_WhenScanning_ThenPageNavigationStopsMutatingSideOrReviewSelectionDirectly(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`applyMoveReviewSelection\(`,
		`applyMoveSideSelection\(`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		return filepath.Base(path) == "cmd_interaction_navigation.go"
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected cmd_interaction_navigation.go page navigation to stop mutating side or review selection directly, actual %v", actualMatches)
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

func TestRefactorGuard_GivenProgramNavigationFile_WhenScanning_ThenRemainingDetailMotionHandlersUseDetailMotionCommands(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`mutateDetailViewStateForYankMotion\(`,
		`mutateDetailViewState\(`,
		`syncCurrentDetailViewport\(`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		return filepath.Base(path) == "program_navigation.go"
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected program_navigation.go remaining detail-motion handlers to dispatch shared detail-motion commands instead of mutating detail shell state inline, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenAuditedDetailMotionShortcutFiles_WhenScanning_ThenTheyDispatchTypedRequestsInsteadOfExecutingDetailMotionCommands(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`program\.executeCmds\(gui, \[]Cmd\{detailMotionCmd\{`,
		`detailMotionCmd\{`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		switch filepath.Base(path) {
		case "program_navigation.go", "pull_request_build_popup_navigation.go", "program_character_motion.go", "pull_request_build_popup_search_repeat.go", "yank_motion.go":
			return true
		default:
			return false
		}
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected the audited detail-motion shortcut files to dispatch typed request messages instead of executing detailMotionCmd directly, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenLegacyDetailMotionShellMutationFiles_WhenScanning_ThenDeadHelpersAreGone(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`func \(program \*Program\) mutateDetailViewStateForYankMotion\(`,
		`func \(program \*Program\) mutatePullRequestBuildRunPopupViewStateWithoutRefresh\(`,
		`func \(program \*Program\) mutatePullRequestBuildRunPopupViewStateForYankMotion\(`,
		`func \(program \*Program\) copySelectedText\(`,
		`func \(program \*Program\) copySelectedPullRequestBuildRunPopupText\(`,
		`func \(program \*Program\) finishPendingYank\(`,
		`func \(program \*Program\) writeTextToClipboard\(`,
		`func \(program \*Program\) copySelectedDetailText\(`,
		`func \(program \*Program\) copySelectedPullRequestURL\(`,
		`func \(program \*Program\) openCurrentLink\(`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		switch filepath.Base(path) {
		case "detail_motion_mutation.go", "yank_motion.go", "yank_url.go", "open_link.go":
			return true
		default:
			return false
		}
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected the dead legacy detail-motion / yank / link helpers to be deleted once the typed command pipeline owns the live flow, actual %v", actualMatches)
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
		"actions_popup_async_cmd.go":               true,
		"assignee_picker_search_cmd.go":            true,
		"cmd_actions_popup_async_requests.go":      true,
		"cmd_detail_fold.go":                       true,
		"cmd_detail_motion.go":                     true,
		"cmd_interaction.go":                       true,
		"cmd_modal_editor_submit_requests.go":      true,
		"cmd_popup_feature_request_requests.go":    true,
		"cmd_popup_feature_requests.go":            true,
		"workflow_command_runtime.go":              true,
		"workflow_session_commands.go":             true,
		"workflow_pull_request_list_commands.go":   true,
		"workflow_pull_request_detail_commands.go": true,
		"workflow_notification_commands.go":        true,
		"workflow_detail_image_commands.go":        true,
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

func TestRefactorGuard_GivenAsyncCommandAndLoaderFiles_WhenScanning_ThenTheyAvoidGuiBoundDispatchAsyncAndLegacyDirectLoaders(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`dispatchAsync\(`,
		`func \(program \*Program\) loadConnectedUser\(`,
		`func \(program \*Program\) loadPullRequests\(`,
		`func \(program \*Program\) loadNotifications\(`,
		`func \(program \*Program\) loadIssueDetail\(`,
		`func \(program \*Program\) loadReleaseDetail\(`,
		`func \(program \*Program\) loadCurrentDetailImageHTML\(`,
		`func \(program \*Program\) loadCurrentDetailImage\(`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		base := filepath.Base(path)
		switch base {
		case "actions_popup_async_cmd.go", "assignee_picker_search_cmd.go", "cmd_interaction_build.go", "cmd_popup_feature_requests.go", "detail_image_loader.go", "dispatch.go", "error_popup.go", "loading_spinner.go", "notification_detail_loader.go", "notification_loading.go", "program_loading.go", "workflow_command_runtime.go", "workflow_session_commands.go", "workflow_pull_request_list_commands.go", "workflow_pull_request_detail_commands.go", "workflow_notification_commands.go", "workflow_detail_image_commands.go":
			return true
		default:
			return false
		}
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected async command and loader files to use the gui-free async message bridge and explicit command surfaces instead of dispatchAsync(...) or direct loader helpers, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenEditorIntentAndModalOpenFiles_WhenScanning_ThenTheyAvoidHandleKeyCallbacksAndFullStateModalPayloads(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`HandleKey\(`,
		`State modalEditorState`,
		`MsgModalEditorOpened\{State:`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		base := filepath.Base(path)
		switch base {
		case "actions_popup_actions.go", "actions_popup_interaction.go", "inline_comment_edit.go", "inline_comment_reply.go", "modal_editor.go", "modal_editor_lifecycle.go", "modal_editor_view.go", "msg_interaction.go", "multiline_editor_input.go", "pull_request_comment_edit.go", "pull_request_custom_search.go", "pull_request_edit.go", "pull_request_review.go", "review_inline_comment.go", "review_submit.go", "search_view.go", "text_input.go", "view_url_prompt.go":
			return true
		default:
			return false
		}
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected editor callback and modal-open files to use typed input intents and open descriptors instead of HandleKey(...) or full modal state payloads, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenRemainingShortcutEntrypointFiles_WhenScanning_ThenTheyDispatchWithoutReducerPreflightOrDirectModalOpen(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`clearPendingSelectionPrefix\(`,
		`detailState\.viewState\.clearPendingPrefix\(`,
		`open(?:Line|Multiline)?ModalEditor(?:WithHeightAndSubmitDescriptor|WithSubmitDescriptor)?\(`,
		`dispatch\(gui,\s*MsgModalEditorOpened\{`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		base := filepath.Base(path)
		return base == "view_url_prompt.go" ||
			base == "pull_request_custom_search.go" ||
			base == "comment_on_pr.go" ||
			base == "review_inline_comment.go" ||
			base == "inline_comment_reply.go" ||
			base == "refresh_active_view.go" ||
			base == "actions_popup_interaction.go" ||
			base == "yank_url.go"
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected the remaining shortcut entrypoint files to stay on dispatch-only behavior while update owns prefixes and modal-open descriptors, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenNotificationShortcutEntrypoints_WhenScanning_ThenTheyDispatchWithoutLocalPreflightOrFeedback(t *testing.T) {
	contents, actualErr := os.ReadFile("notification_actions.go")
	if actualErr != nil {
		t.Fatalf("read notification_actions.go: %v", actualErr)
	}

	forbiddenPatterns := map[string]*regexp.Regexp{
		"markSelectedNotificationRead":       regexp.MustCompile(`func \(program \*Program\) markSelectedNotificationRead\(gui \*gocui\.Gui\) error \{[^}]*selectedNotificationActionTarget\(`),
		"markSelectedNotificationDone":       regexp.MustCompile(`func \(program \*Program\) markSelectedNotificationDone\(gui \*gocui\.Gui\) error \{[^}]*selectedNotificationActionTarget\(`),
		"openSelectedNotificationInBrowser":  regexp.MustCompile(`func \(program \*Program\) openSelectedNotificationInBrowser\(gui \*gocui\.Gui\) error \{[^}]*(?:selectedNotificationBrowserURL\(|ErrLinkOpenerUnavailable|MsgOpenBrowserURLRequested\{)`),
		"handleNotificationKeyAction":        regexp.MustCompile(`func \(program \*Program\) handleNotificationKeyAction\(`),
		"directNotificationFeedbackDispatch": regexp.MustCompile(`dispatch\(gui, MsgFeedbackSet\{Target: program\.model\.Focus\(\), Message: err\.Error\(\)\}\)`),
	}

	for name, pattern := range forbiddenPatterns {
		if pattern.Match(contents) {
			t.Fatalf("expected notification shortcut entrypoints to avoid local preflight or feedback in %s", name)
		}
	}
}

func TestRefactorGuard_GivenDetailMotionAndNavigationCommandFiles_WhenScanning_ThenTheyDoNotMutateDurableDetailOrBuildPopupState(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`program\.detailState\.viewState\.`,
		`program\.pullRequestBuildRunPopup\.viewState\.`,
		`popup\.viewState\.`,
		`mutateDetailViewState`,
		`mutatePullRequestBuildRunPopupViewState`,
		`syncDetailViewState\(`,
		`focusDetailLine\(`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		base := filepath.Base(path)
		return base == "cmd_detail_motion.go" || base == "cmd_interaction_navigation.go" || base == "cmd_interaction_detail_search.go"
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected detail-motion and navigation command executors to resolve live context then dispatch typed reducer messages instead of mutating durable detail/build-popup state directly, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenProductionFiles_WhenScanning_ThenOnlyUpdateAndCacheApplyFilesWritePullRequestDetailOrDiffCaches(t *testing.T) {
	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", regexp.MustCompile(`program\.(?:pullRequestDetailCache|pullRequestDiffCache)\[[^]]+\]\s*=`), func(path string) bool {
		base := filepath.Base(path)
		return strings.HasSuffix(base, ".go") && !strings.HasSuffix(base, "_test.go")
	})

	remainingMatches := make([]string, 0, len(actualMatches))
	for _, match := range actualMatches {
		base := filepath.Base(strings.Split(match, ":")[0])
		if strings.HasPrefix(base, "update") || base == "detail_image_html_apply.go" || base == "pull_request_cache_apply.go" {
			continue
		}
		remainingMatches = append(remainingMatches, match)
	}
	if len(remainingMatches) != 0 {
		t.Fatalf("expected pull-request detail and diff cache writes to stay in update or dedicated cache-apply files, actual %v", remainingMatches)
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

type guardSourceFile struct {
	path         string
	relativePath string
	lines        []string
}

var (
	guardSourceFilesOnce sync.Once
	guardSourceFiles     []guardSourceFile
	guardSourceFilesErr  error
)

func given_regexpLineMatchesInGoFiles(t *testing.T, root string, pattern *regexp.Regexp, include func(string) bool) []string {
	t.Helper()

	scanRoot := given_guardScanRoot(t, root)
	actualMatches := make([]string, 0)
	for _, file := range given_guardSourceFiles(t) {
		if !given_guardFileUnderRoot(file.path, scanRoot) {
			continue
		}
		if include != nil && !include(file.path) {
			continue
		}
		for lineIndex, line := range file.lines {
			if !pattern.MatchString(line) {
				continue
			}
			actualMatches = append(actualMatches, fmt.Sprintf("%s:%d contains %q", file.relativePath, lineIndex+1, strings.TrimSpace(line)))
		}
	}

	slices.Sort(actualMatches)
	return actualMatches
}

func given_guardSourceFiles(t *testing.T) []guardSourceFile {
	t.Helper()

	packageRoot := given_guardPackageRoot(t)
	guardSourceFilesOnce.Do(func() {
		guardSourceFiles = make([]guardSourceFile, 0)
		guardSourceFilesErr = filepath.WalkDir(packageRoot, func(path string, entry os.DirEntry, walkErr error) error {
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
			if !strings.HasSuffix(entry.Name(), ".go") {
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
			guardSourceFiles = append(guardSourceFiles, guardSourceFile{
				path:         path,
				relativePath: filepath.ToSlash(relativePath),
				lines:        strings.Split(string(contents), "\n"),
			})
			return nil
		})
	})
	if guardSourceFilesErr != nil {
		t.Fatalf("walk go files: %v", guardSourceFilesErr)
	}
	return guardSourceFiles
}

func given_guardFileUnderRoot(path string, scanRoot string) bool {
	if path == scanRoot {
		return true
	}
	return strings.HasPrefix(path, scanRoot+string(os.PathSeparator))
}
