//go:build manualvisual

package tui

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jesseduffield/gocui"
)

const (
	commandHistoryManualVisualReadyTokenEnv = "LAZYGH_TMUX_READY_TOKEN"
	commandHistoryManualVisualDoneTokenEnv  = "LAZYGH_TMUX_DONE_TOKEN"
)

func TestManualVisual_CommandHistoryPopup(t *testing.T) {
	readyToken := strings.TrimSpace(os.Getenv(commandHistoryManualVisualReadyTokenEnv))
	doneToken := strings.TrimSpace(os.Getenv(commandHistoryManualVisualDoneTokenEnv))
	if readyToken == "" || doneToken == "" {
		t.Skip("manualvisual command-history check needs tmux wait-for tokens")
	}

	subject := NewProgramWithModel(given_model())
	later := time.Date(2026, time.September, 13, 12, 35, 2, 0, time.Local)
	Update(subject, MsgGHCommandStarted{Command: "gh pr view 42 -R acme/widgets --web", StartedAt: later})
	Update(subject, MsgGHCommandStarted{Command: "gh api graphql", StartedAt: later.Add(-6 * time.Second)})
	Update(subject, MsgOpenCommandMode{})
	for _, character := range "history" {
		Update(subject, MsgCommandInputRequested{Intent: newLineEditorInsertRuneIntent(character)})
	}
	Update(subject, MsgSubmitCommand{})

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
	if actualErr = subject.layout(gui); actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}

	stopPolling := make(chan struct{})
	pollingStopped := make(chan struct{})
	go func() {
		defer close(pollingStopped)
		if actualErr := runCommandHistoryManualVisualSequence(t, gui, subject, readyToken, stopPolling); actualErr != nil {
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

func runCommandHistoryManualVisualSequence(t *testing.T, gui *gocui.Gui, subject *Program, readyToken string, stop <-chan struct{}) error {
	t.Helper()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return nil
		case <-ticker.C:
			ready := make(chan bool, 1)
			errCh := make(chan error, 1)
			gui.Update(func(gui *gocui.Gui) error {
				popupView, actualErr := gui.View(viewPullRequestBuildInfoName)
				if actualErr != nil {
					ready <- false
					errCh <- nil
					return nil
				}
				if popupView.Title != "history" || !strings.Contains(popupView.Buffer(), "12:34:56 gh api graphql") || !strings.Contains(popupView.Buffer(), "12:35:02 gh pr view 42 -R acme/widgets --web") {
					ready <- false
					errCh <- nil
					return nil
				}
				currentView := gui.CurrentView()
				if currentView == nil || currentView.Name() != viewPullRequestBuildInfoName {
					ready <- false
					errCh <- nil
					return nil
				}

				bindings := subject.keybindingSpecs()
				if actualErr := given_handlerForBinding(t, bindings, viewPullRequestBuildInfoName, 'j')(gui, popupView); actualErr != nil {
					ready <- false
					errCh <- actualErr
					return actualErr
				}
				if actualErr := given_handlerForBinding(t, bindings, viewPullRequestBuildInfoName, '/')(gui, popupView); actualErr != nil {
					ready <- false
					errCh <- actualErr
					return actualErr
				}
				searchView, actualErr := gui.View(viewSearchName)
				if actualErr != nil {
					ready <- false
					errCh <- actualErr
					return actualErr
				}
				for _, character := range "gh api" {
					if !subject.editSearch(searchView, 0, character, gocui.ModNone) {
						actualErr = errors.New("expected command-history popup search input to be handled")
						ready <- false
						errCh <- actualErr
						return actualErr
					}
				}
				if actualErr := given_handlerForBinding(t, subject.keybindingSpecs(), viewSearchName, gocui.KeyEnter)(gui, searchView); actualErr != nil {
					ready <- false
					errCh <- actualErr
					return actualErr
				}
				popupView, actualErr = gui.View(viewPullRequestBuildInfoName)
				if actualErr != nil || subject.pullRequestBuildRunPopup.searchQuery != "gh api" {
					if actualErr == nil {
						actualErr = errors.New("expected command-history popup search query")
					}
					ready <- false
					errCh <- actualErr
					return actualErr
				}
				if actualErr := given_handlerForBinding(t, subject.keybindingSpecs(), viewPullRequestBuildInfoName, gocui.KeyEsc)(gui, popupView); actualErr != nil {
					ready <- false
					errCh <- actualErr
					return actualErr
				}
				if subject.pullRequestBuildRunPopup != nil || subject.commandModeActive() {
					ready <- false
					errCh <- nil
					return nil
				}

				subject.commandHistoryStore = newCommandHistoryStore()
				if actualErr := subject.dispatch(gui, MsgOpenCommandMode{}); actualErr != nil {
					ready <- false
					errCh <- actualErr
					return actualErr
				}
				for _, character := range "history" {
					if actualErr := subject.dispatch(gui, MsgCommandInputRequested{Intent: newLineEditorInsertRuneIntent(character)}); actualErr != nil {
						ready <- false
						errCh <- actualErr
						return actualErr
					}
				}
				if actualErr := subject.dispatch(gui, MsgSubmitCommand{}); actualErr != nil {
					ready <- false
					errCh <- actualErr
					return actualErr
				}
				emptyPopupView, actualErr := gui.View(viewPullRequestBuildInfoName)
				if actualErr != nil || strings.TrimSpace(emptyPopupView.Buffer()) != "" {
					if actualErr == nil {
						actualErr = errors.New("expected empty command-history popup body")
					}
					ready <- false
					errCh <- actualErr
					return actualErr
				}

				ready <- true
				errCh <- signalTmuxWaitToken(readyToken)
				return nil
			})
			if actualErr := <-errCh; actualErr != nil {
				return actualErr
			}
			if !<-ready {
				continue
			}
			return nil
		}
	}
}
