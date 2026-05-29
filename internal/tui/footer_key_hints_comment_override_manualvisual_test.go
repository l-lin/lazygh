//go:build manualvisual

package tui

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"

	appconfig "github.com/l-lin/lazygh/internal/config"
	"github.com/l-lin/lazygh/internal/githubcli"
)

const (
	footerCommentHintReadyTokenEnv = "LAZYGH_TMUX_READY_TOKEN"
	footerCommentHintDoneTokenEnv  = "LAZYGH_TMUX_DONE_TOKEN"
)

func TestManualVisual_FooterKeyHintsShowCommentOverrideInBrowserChangesDetail(t *testing.T) {
	readyToken := strings.TrimSpace(os.Getenv(footerCommentHintReadyTokenEnv))
	doneToken := strings.TrimSpace(os.Getenv(footerCommentHintDoneTokenEnv))
	if readyToken == "" || doneToken == "" {
		t.Skip("manualvisual footer comment-hint check needs tmux wait-for tokens")
	}

	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": given_pullRequestDetailForChangesInlineCommentTests()},
		diffs:   map[string]githubcli.PullRequestDiff{"acme/widgets#42": given_reviewSessionPullRequestDiff()},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.ApplyKeymapOverrides(appconfig.KeymapOverrides{
		"global": {
			"open_actions_popup": {"alt+a"},
		},
		"pull_requests": {
			"comment_on_pull_request": {"alt+c"},
		},
	})
	gui, actualErr := gocui.NewGui(gocui.NewGuiOpts{OutputMode: gocui.OutputTrue})
	if actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}
	defer gui.Close()

	subject.configureGUI(gui)
	gui.SetManagerFunc(subject.layout)
	if actualErr = subject.setKeybindings(gui); actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}
	given_browserChangesDetailFocusForInlineComment(t, gui, subject)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "new line")
	if actualErr = subject.afterStateChange(gui); actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}

	stopPolling := make(chan struct{})
	pollingStopped := make(chan struct{})
	go func() {
		defer close(pollingStopped)
		if actualErr := signalTmuxWaitTokenWhenCondition(gui, readyToken, stopPolling, func(gui *gocui.Gui) bool {
			keyHintsView, actualErr := gui.View(viewStatusLineKeyHintsName)
			if actualErr != nil {
				return false
			}
			return strings.Contains(strings.TrimSpace(keyHintsView.Buffer()), "Alt+C: comment, Alt+A: action")
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
