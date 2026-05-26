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
	buildPopupManualVisualReadyTokenEnv = "LAZYGH_TMUX_READY_TOKEN"
	buildPopupManualVisualDoneTokenEnv  = "LAZYGH_TMUX_DONE_TOKEN"
)

func TestManualVisual_BuildPopupSearchNavigation(t *testing.T) {
	readyToken := strings.TrimSpace(os.Getenv(buildPopupManualVisualReadyTokenEnv))
	doneToken := strings.TrimSpace(os.Getenv(buildPopupManualVisualDoneTokenEnv))
	if readyToken == "" || doneToken == "" {
		t.Skip("manualvisual build-popup check needs tmux wait-for tokens")
	}

	model := given_pullRequestCommentModel()
	model.OpenDetail()
	subject := given_pullRequestCommentProgram(model, &fakePullRequestDetailLoader{})
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
	if actualErr = subject.openPullRequestBuildRunPopup(gui, pullRequestBuildRunPopupContent{
		checkTitle: "CI / test",
		body: strings.Join([]string{
			"Line 1",
			"Line 2",
			"Line 3",
			"Line 4",
			"Line 5",
			"Target first",
			"Line 7",
			"Line 8",
			"Target second",
			"Line 10",
		}, "\n"),
	}); actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}

	stopPolling := make(chan struct{})
	pollingStopped := make(chan struct{})
	go func() {
		defer close(pollingStopped)
		if actualErr := runBuildPopupManualVisualSequence(t, gui, subject, readyToken, stopPolling); actualErr != nil {
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

func TestManualVisual_BuildPopupCharacterMotionSelection(t *testing.T) {
	readyToken := strings.TrimSpace(os.Getenv(buildPopupManualVisualReadyTokenEnv))
	doneToken := strings.TrimSpace(os.Getenv(buildPopupManualVisualDoneTokenEnv))
	if readyToken == "" || doneToken == "" {
		t.Skip("manualvisual build-popup check needs tmux wait-for tokens")
	}

	model := given_pullRequestCommentModel()
	model.OpenDetail()
	subject := given_pullRequestCommentProgram(model, &fakePullRequestDetailLoader{})
	subject.clipboardWriter = &fakeClipboardWriter{}
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
	if actualErr = subject.openPullRequestBuildRunPopup(gui, pullRequestBuildRunPopupContent{checkTitle: "CI / test", body: "banana"}); actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}

	stopPolling := make(chan struct{})
	pollingStopped := make(chan struct{})
	go func() {
		defer close(pollingStopped)
		if actualErr := runBuildPopupCharacterMotionManualVisualSequence(t, gui, subject, readyToken, stopPolling); actualErr != nil {
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

func runBuildPopupManualVisualSequence(t *testing.T, gui *gocui.Gui, subject *Program, readyToken string, stop <-chan struct{}) error {
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

				bindings := subject.keybindingSpecs()
				if actualErr := given_handlerForBinding(t, bindings, viewPullRequestBuildInfoName, 'j')(gui, popupView); actualErr != nil {
					ready <- false
					errCh <- actualErr
					return actualErr
				}
				if actualErr := given_handlerForBinding(t, bindings, viewPullRequestBuildInfoName, gocui.KeyCtrlD)(gui, popupView); actualErr != nil {
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
				for _, character := range "Target" {
					if actual := subject.editSearch(searchView, 0, character, gocui.ModNone); !actual {
						ready <- false
						errCh <- errors.New("expected popup search typing to be handled")
						return nil
					}
				}
				if actualErr := given_handlerForBinding(t, subject.keybindingSpecs(), viewSearchName, gocui.KeyEnter)(gui, searchView); actualErr != nil {
					ready <- false
					errCh <- actualErr
					return actualErr
				}

				popupView, actualErr = gui.View(viewPullRequestBuildInfoName)
				if actualErr != nil {
					ready <- false
					errCh <- actualErr
					return actualErr
				}
				if actualErr := given_handlerForBinding(t, subject.keybindingSpecs(), viewPullRequestBuildInfoName, 'n')(gui, popupView); actualErr != nil {
					ready <- false
					errCh <- actualErr
					return actualErr
				}

				currentView := gui.CurrentView()
				if currentView == nil || currentView.Name() != viewPullRequestBuildInfoName {
					ready <- false
					errCh <- nil
					return nil
				}
				if !strings.Contains(popupView.Buffer(), "Target second") {
					ready <- false
					errCh <- nil
					return nil
				}

				ready <- true
				errCh <- nil
				return nil
			})
			if actualErr := <-errCh; actualErr != nil {
				return actualErr
			}
			if !<-ready {
				continue
			}
			return signalTmuxWaitToken(readyToken)
		}
	}
}

func runBuildPopupCharacterMotionManualVisualSequence(t *testing.T, gui *gocui.Gui, subject *Program, readyToken string, stop <-chan struct{}) error {
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

				bindings := subject.registeredKeybindingSpecs()
				for _, key := range []rune{'v', 'f'} {
					if actualErr := given_handlerForBinding(t, bindings, viewPullRequestBuildInfoName, key)(gui, popupView); actualErr != nil {
						ready <- false
						errCh <- actualErr
						return actualErr
					}
				}
				bindings = subject.registeredKeybindingSpecs()
				if actualErr := given_handlerForBinding(t, bindings, viewPullRequestBuildInfoName, 'a')(gui, popupView); actualErr != nil {
					ready <- false
					errCh <- actualErr
					return actualErr
				}
				bindings = subject.registeredKeybindingSpecs()
				if actualErr := given_handlerForBinding(t, bindings, viewPullRequestBuildInfoName, 'y')(gui, popupView); actualErr != nil {
					ready <- false
					errCh <- actualErr
					return actualErr
				}

				clipboardWriter, ok := subject.clipboardWriter.(*fakeClipboardWriter)
				if !ok || len(clipboardWriter.writes) != 1 || clipboardWriter.writes[0] != "ba" {
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

				ready <- true
				errCh <- nil
				return nil
			})
			if actualErr := <-errCh; actualErr != nil {
				return actualErr
			}
			if !<-ready {
				continue
			}
			return signalTmuxWaitToken(readyToken)
		}
	}
}
