package tui

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func given_guardPackageRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve guard test file path")
	}
	return filepath.Dir(file)
}

func given_guardScanRoot(t *testing.T, root string) string {
	t.Helper()

	if filepath.IsAbs(root) {
		return filepath.Clean(root)
	}
	return filepath.Clean(filepath.Join(given_guardPackageRoot(t), root))
}

func TestRefactorGuard_GivenGuardPackageRoot_WhenResolvingTheTUIRoot_ThenItFindsProgramGo(t *testing.T) {
	_, actualErr := os.Stat(filepath.Join(given_guardPackageRoot(t), "program.go"))
	if actualErr != nil {
		t.Fatalf("expected to resolve the TUI source directory, actual error %v", actualErr)
	}
}
