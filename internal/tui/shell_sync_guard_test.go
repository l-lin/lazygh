package tui

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestRefactorGuard_GivenProductionFiles_WhenScanning_ThenKeybindingReloadLivesInShellSyncOnly(t *testing.T) {
	allowedFiles := map[string]bool{
		"program_keybindings.go": true,
		"shell_sync.go":          true,
	}

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", regexp.MustCompile(`reloadRegisteredKeybindings\(`), func(path string) bool {
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
		t.Fatalf("expected keybinding reload to stay confined to shell sync, actual %v", remainingMatches)
	}
}

func TestRenderEntryPoints_GivenRenderGo_WhenInspecting_ThenItDoesNotOwnStartupOrKeybindingSync(t *testing.T) {
	contents, actualErr := os.ReadFile("render.go")
	then_noError(t, actualErr)
	actualSource := string(contents)

	for _, forbiddenSnippet := range []string{"MsgAppStarted", "reloadRegisteredKeybindings(gui)"} {
		if strings.Contains(actualSource, forbiddenSnippet) {
			t.Fatalf("expected render.go to avoid %q, actual source:\n%s", forbiddenSnippet, actualSource)
		}
	}
}

func TestProgramStart_GivenFreshProgram_WhenStarting_ThenItMarksStartedAndSyncsTheCurrentView(t *testing.T) {
	subject := NewProgramWithModel(defaultProgramModel())
	gui := given_headlessGui(t)
	defer gui.Close()

	actualErr := subject.start(gui)
	then_noError(t, actualErr)

	if !subject.startupState.appStarted {
		t.Fatal("expected the program to be marked as started during startup")
	}
	if actual := subject.navigationState.registeredKeybindingFingerprint; actual == "" {
		t.Fatal("expected startup to register the initial keybindings")
	}
	then_currentViewNameIs(t, gui, viewPullRequestsName)
}
