//go:build manualvisual

package tui

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"

	"github.com/l-lin/lazygh/internal/githubcli"
)

const (
	replyCodeFenceManualVisualReadyTokenEnv = "LAZYGH_TMUX_READY_TOKEN"
	replyCodeFenceManualVisualDoneTokenEnv  = "LAZYGH_TMUX_DONE_TOKEN"
)

func TestManualVisual_BrowserChangesInlineThreadReplyCodeFenceSpacing(t *testing.T) {
	readyToken := strings.TrimSpace(os.Getenv(replyCodeFenceManualVisualReadyTokenEnv))
	doneToken := strings.TrimSpace(os.Getenv(replyCodeFenceManualVisualDoneTokenEnv))
	if readyToken == "" || doneToken == "" {
		t.Skip("manualvisual reply code-fence check needs tmux wait-for tokens")
	}

	diff := githubcli.PullRequestDiff{UnifiedDiff: "diff --git a/internal/tui/render.go b/internal/tui/render.go\nindex 0000000..1111111 100644\n--- a/internal/tui/render.go\n+++ b/internal/tui/render.go\n@@ -42,0 +43,3 @@\n+func render(value int) string {\n+\treturn fmt.Sprintf(\"%d\", value + 42)\n+}\n"}
	diff.Threads = []githubcli.PullRequestReviewThread{{
		ID:       "thread-1",
		Path:     "internal/tui/render.go",
		Line:     45,
		DiffSide: "RIGHT",
		Comments: []githubcli.PullRequestComment{
			{
				Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"},
				Body:      "Root comment",
				CreatedAt: "2026-04-18T10:00:00Z",
				DiffHunk:  "@@ -42,0 +43,3 @@\n+func render(value int) string {\n+\treturn fmt.Sprintf(\"%d\", value + 42)\n+}",
			},
			{
				Author:    &githubcli.PullRequestCommentAuthor{Login: "octocat"},
				Body:      "```go\nfunc render(value int) string {\n\treturn fmt.Sprintf(\"%d\", value + 42)\n}\n```",
				CreatedAt: "2026-04-18T10:30:00Z",
			},
		},
	}}
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "First PR",
				Number:      42,
				Body:        "Body 42",
				BaseRefName: "main",
				HeadRefName: "feature/changes",
				State:       "OPEN",
			},
		},
		diffs: map[string]githubcli.PullRequestDiff{"acme/widgets#42": diff},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)

	gui, actualErr := gocui.NewGui(gocui.NewGuiOpts{OutputMode: gocui.OutputTrue})
	if actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}
	defer gui.Close()

	subject.configureGUI(gui)
	gui.SetManagerFunc(subject.layout)
	if actualErr = subject.layout(gui); actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}
	if actualErr = subject.setKeybindings(gui); actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}
	if actualErr = subject.openDetail(gui, nil); actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}
	subject.detailState.activeTab = ChangesDetailTab
	if actualErr = subject.afterStateChange(gui); actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}

	stopPolling := make(chan struct{})
	pollingStopped := make(chan struct{})
	go func() {
		defer close(pollingStopped)
		if actualErr := signalTmuxWaitTokenWhenCondition(gui, readyToken, stopPolling, func(gui *gocui.Gui) bool {
			detailView, detailErr := gui.View(viewDetailName)
			if detailErr != nil {
				return false
			}
			buffer := detailView.Buffer()
			return strings.Contains(buffer, "@octocat") && strings.Contains(buffer, "func render(value int) string {")
		}); actualErr != nil {
			return
		}
		if actualErr := waitForTmuxToken(doneToken); actualErr != nil {
			return
		}
		gui.Update(func(*gocui.Gui) error {
			return gocui.ErrQuit
		})
	}()

	actualErr = gui.MainLoop()
	close(stopPolling)
	<-pollingStopped
	if actualErr != nil && !errors.Is(actualErr, gocui.ErrQuit) {
		t.Fatalf("expected no error, actual %v", actualErr)
	}
}
