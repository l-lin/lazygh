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

func TestRefactorGuard_GivenProductionFiles_WhenScanning_ThenDirectFeedbackAndErrorReportingStayInUpdateFiles(t *testing.T) {
	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", regexp.MustCompile(`program\.(?:setFeedback|reportError)\(`), func(path string) bool {
		base := filepath.Base(path)
		return strings.HasSuffix(base, ".go") && !strings.HasSuffix(base, "_test.go")
	})

	remainingMatches := make([]string, 0, len(actualMatches))
	for _, match := range actualMatches {
		base := filepath.Base(strings.Split(match, ":")[0])
		if strings.HasPrefix(base, "update") {
			continue
		}
		remainingMatches = append(remainingMatches, match)
	}
	if len(remainingMatches) != 0 {
		t.Fatalf("expected direct feedback and error reporting to stay in update files, actual %v", remainingMatches)
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
