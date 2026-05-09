package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReleaseProjectFiles_GivenMiseToml_WhenReadingTheTasks_ThenItDefinesGoReleaserValidationAndSnapshotTasks(t *testing.T) {
	contents, actualErr := os.ReadFile(filepath.Join("..", "..", "mise.toml"))
	then_noError(t, actualErr)

	actual := string(contents)
	for _, expected := range []string{
		"[tasks.release-check]",
		"ghcr.io/goreleaser/goreleaser-cross:v1.25.9",
		"check --config .goreleaser.yaml",
		"[tasks.release-snapshot]",
		"release --snapshot --clean --skip=publish --config .goreleaser.yaml",
	} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("expected mise.toml to contain %q, actual %q", expected, actual)
		}
	}
}

func TestReleaseProjectFiles_GivenTheRepo_WhenInspectingTheGoReleaserConfig_ThenItDefinesCrossPlatformArchivesForGitHubAndCodeberg(t *testing.T) {
	contents, actualErr := os.ReadFile(filepath.Join("..", "..", ".goreleaser.yaml"))
	then_noError(t, actualErr)

	actual := string(contents)
	for _, expected := range []string{
		"version: 2",
		"project_name: lazygh",
		"gitea_urls:",
		"api: https://codeberg.org/api/v1",
		"download: https://codeberg.org",
		"CGO_ENABLED=1",
		"CC_darwin_amd64=o64-clang",
		"CC_darwin_arm64=oa64-clang",
		"CC_linux_amd64=x86_64-linux-gnu-gcc",
		"CC_linux_arm64=aarch64-linux-gnu-gcc",
		"CC_windows_amd64=x86_64-w64-mingw32-gcc",
		"ignore:",
		"goos: windows",
		"goarch: arm64",
		"goos:",
		"- linux",
		"- darwin",
		"- windows",
		"goarch:",
		"- amd64",
		"- arm64",
		"format_overrides:",
		"goos: windows",
		"formats: [zip]",
		"checksum:",
		"name_template: checksums.txt",
	} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("expected .goreleaser.yaml to contain %q, actual %q", expected, actual)
		}
	}
}

func TestReleaseProjectFiles_GivenTheRepo_WhenInspectingTheReleaseWorkflows_ThenItDefinesGitHubAndCodebergTaggedReleaseJobs(t *testing.T) {
	githubContents, actualErr := os.ReadFile(filepath.Join("..", "..", ".github", "workflows", "release.yml"))
	then_noError(t, actualErr)

	githubWorkflow := string(githubContents)
	for _, expected := range []string{
		"tags:",
		"- 'v*'",
		"runs-on: ubuntu-latest",
		"image: ghcr.io/goreleaser/goreleaser-cross:v1.25.9",
		"git remote add origin \"${GITHUB_SERVER_URL}/${GITHUB_REPOSITORY}.git\"",
		"go test ./...",
		"go build -o ./bin/lazygh ./cmd/lazygh",
		"goreleaser release --clean --config .goreleaser.yaml",
		"GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}",
	} {
		if !strings.Contains(githubWorkflow, expected) {
			t.Fatalf("expected .github/workflows/release.yml to contain %q, actual %q", expected, githubWorkflow)
		}
	}

	codebergContents, actualErr := os.ReadFile(filepath.Join("..", "..", ".forgejo", "workflows", "release.yml"))
	then_noError(t, actualErr)

	codebergWorkflow := string(codebergContents)
	for _, expected := range []string{
		"tags:",
		"- 'v*'",
		"runs-on: codeberg-medium",
		"image: ghcr.io/goreleaser/goreleaser-cross:v1.25.9",
		"git remote add origin \"${FORGEJO_SERVER_URL}/${FORGEJO_REPOSITORY}.git\"",
		"go test ./...",
		"go build -o ./bin/lazygh ./cmd/lazygh",
		"goreleaser release --clean --config .goreleaser.yaml",
		"GORELEASER_FORCE_TOKEN: gitea",
		"GITEA_TOKEN: ${{ secrets.GITHUB_TOKEN }}",
	} {
		if !strings.Contains(codebergWorkflow, expected) {
			t.Fatalf("expected .forgejo/workflows/release.yml to contain %q, actual %q", expected, codebergWorkflow)
		}
	}
}

func TestReleaseProjectFiles_GivenTheReadme_WhenReadingTheReleaseSection_ThenItDocumentsGoReleaserTasksAndTaggedPublishing(t *testing.T) {
	contents, actualErr := os.ReadFile(filepath.Join("..", "..", "README.md"))
	then_noError(t, actualErr)

	actual := string(contents)
	for _, expected := range []string{
		"mise run release-check",
		"mise run release-snapshot",
		"Tagged pushes that match `v*` publish release archives",
		"GitHub runs `.github/workflows/release.yml`",
		"Codeberg runs `.forgejo/workflows/release.yml`",
		"`linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`, and `windows/amd64`",
		"`ghcr.io/goreleaser/goreleaser-cross:v1.25.9`",
	} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("expected README.md to contain %q, actual %q", expected, actual)
		}
	}
}
