//go:build manualvisual

package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jesseduffield/gocui"

	"github.com/l-lin/lazygh/internal/githubcli"
)

const (
	reviewInlineCommentExternalEditorReadyTokenEnv = "LAZYGH_TMUX_READY_TOKEN"
	reviewInlineCommentExternalEditorDoneTokenEnv  = "LAZYGH_TMUX_DONE_TOKEN"
)

func TestManualVisual_ReviewInlineCommentExternalEditorReturnsToTheModal(t *testing.T) {
	readyToken := strings.TrimSpace(os.Getenv(reviewInlineCommentExternalEditorReadyTokenEnv))
	doneToken := strings.TrimSpace(os.Getenv(reviewInlineCommentExternalEditorDoneTokenEnv))
	if readyToken == "" || doneToken == "" {
		t.Skip("manualvisual inline-comment external-editor check needs tmux wait-for tokens")
	}

	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/widgets#42": given_reviewSessionPullRequestDiff(),
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	nvimPath, actualErr := exec.LookPath("nvim")
	if actualErr != nil {
		t.Skip("manualvisual inline-comment external-editor check needs nvim")
	}
	editorScriptPath := filepath.Join(t.TempDir(), "fake-editor.sh")
	editorScript := []byte("#!/bin/sh\nsleep 1\nexec \"" + nvimPath + "\" -u NONE -i NONE -n '+set nomore shortmess+=I' '+silent! %delete _' '+call setline(1, [\"Edited in external editor\"])' '+wq' \"$1\"\n")
	if actualErr := os.WriteFile(editorScriptPath, editorScript, 0o755); actualErr != nil {
		t.Fatalf("expected no error writing editor shim, actual %v", actualErr)
	}
	t.Setenv("EDITOR", editorScriptPath)
	subject.externalEditor = systemExternalEditor{}

	gui, actualErr := gocui.NewGui(gocui.NewGuiOpts{OutputMode: gocui.OutputTrue})
	if actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}
	defer gui.Close()
	stopLoadingSpinner := subject.startLoadingSpinner(gui)
	defer stopLoadingSpinner()

	subject.configureGUI(gui)
	gui.SetManagerFunc(subject.layout)
	if actualErr = subject.setKeybindings(gui); actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}
	if actualErr = subject.layout(gui); actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}
	if actualErr = given_startingReviewMode(t, gui, subject); actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}
	if actualErr = subject.focusDetailView(gui, nil); actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "new line")

	detailView, actualErr := gui.View(viewDetailName)
	if actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}
	if actualErr = given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, 'c')(gui, detailView); actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}
	subject.overlayState.modalEditor.editor.SetText("Draft inline comment")
	if actualErr = subject.afterStateChange(gui); actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}

	stopPolling := make(chan struct{})
	pollingStopped := make(chan struct{})
	pollingErrCh := make(chan error, 1)
	go func() {
		defer close(pollingStopped)
		if actualErr := runReviewInlineCommentExternalEditorManualVisualSequence(t, gui, subject, readyToken, stopPolling); actualErr != nil {
			pollingErrCh <- actualErr
			gui.Update(func(*gocui.Gui) error {
				return gocui.ErrQuit
			})
			return
		}
		if actualErr := waitForTmuxToken(doneToken); actualErr != nil {
			pollingErrCh <- actualErr
			gui.Update(func(*gocui.Gui) error {
				return gocui.ErrQuit
			})
			return
		}
		gui.Update(func(*gocui.Gui) error {
			return gocui.ErrQuit
		})
	}()

	actualErr = gui.MainLoop()
	close(stopPolling)
	<-pollingStopped
	select {
	case pollingErr := <-pollingErrCh:
		t.Fatalf("expected no error, actual %v", pollingErr)
	default:
	}
	if actualErr != nil && actualErr != gocui.ErrQuit {
		t.Fatalf("expected no error, actual %v", actualErr)
	}
}

func runReviewInlineCommentExternalEditorManualVisualSequence(t *testing.T, gui *gocui.Gui, subject *Program, readyToken string, stop <-chan struct{}) error {
	t.Helper()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	deadline := time.Now().Add(4 * time.Second)
	externalEditTriggered := false

	for {
		select {
		case <-stop:
			return nil
		case <-ticker.C:
			if time.Now().After(deadline) {
				return fmt.Errorf("timed out waiting for the external editor to return to the modal; last snapshot=%q", gui.Snapshot())
			}

			readyCh := make(chan bool, 1)
			errCh := make(chan error, 1)
			gui.Update(func(gui *gocui.Gui) error {
				modalView, actualErr := gui.View(viewModalEditorName)
				if actualErr != nil {
					readyCh <- false
					errCh <- nil
					return nil
				}
				currentView := gui.CurrentView()
				if currentView == nil || currentView.Name() != viewModalEditorName {
					readyCh <- false
					errCh <- nil
					return nil
				}

				if !externalEditTriggered {
					if !strings.Contains(modalView.Buffer(), "Draft inline comment") {
						readyCh <- false
						errCh <- nil
						return nil
					}
					handler := given_handlerForBinding(t, subject.keybindingSpecs(), viewModalEditorName, gocui.KeyCtrlG)
					if actualErr = handler(gui, modalView); actualErr != nil {
						readyCh <- false
						errCh <- actualErr
						return actualErr
					}
					externalEditTriggered = true
					readyCh <- false
					errCh <- nil
					return nil
				}

				snapshot := gui.Snapshot()
				if !strings.Contains(snapshot, pullRequestReviewInlineCommentComposerTitle) || !strings.Contains(snapshot, "Edited in external editor") {
					readyCh <- false
					errCh <- nil
					return nil
				}
				if !strings.Contains(modalView.Buffer(), "Edited in external editor") {
					readyCh <- false
					errCh <- fmt.Errorf("expected modal buffer to contain edited text, actual %q", modalView.Buffer())
					return nil
				}

				readyCh <- true
				errCh <- nil
				return nil
			})
			if actualErr := <-errCh; actualErr != nil {
				return actualErr
			}
			if !<-readyCh {
				continue
			}
			return signalTmuxWaitToken(readyToken)
		}
	}
}
