//go:build manualvisual

package tui

import (
	"fmt"
	"os"
	"testing"

	"github.com/jesseduffield/gocui"
)

func TestGreySyntax_GivenLiveTTY_WhenRendering_ThenDiffBackgroundsKeepPriority(t *testing.T) {
	ready := os.Getenv("LAZYGH_GREY_READY_TOKEN")
	done := os.Getenv("LAZYGH_GREY_DONE_TOKEN")
	if ready == "" || done == "" {
		t.Skip("live grey verification needs tmux wait-for tokens")
	}
	given_greySyntaxPalette(t)
	file := reviewDiffFile{
		Path: "main.go",
		Hunks: []reviewDiffHunk{{
			Header: "@@ -1,3 +1,3 @@",
			Lines: []reviewDiffLine{
				{Kind: reviewDiffContextLine, Text: `func sample() { sample("hello", 42) }`, LeftLine: 1, RightLine: 1},
				{Kind: reviewDiffDeletionLine, Text: `func before() { sample("old", 10) }`, LeftLine: 2},
				{Kind: reviewDiffAdditionLine, Text: `func after() { sample("new", 20) }`, RightLine: 2},
			},
		}},
	}
	renderer := glamourMarkdownRenderer{}
	markdown, actualErr := renderer.Render("```go\nfunc markdownSample() { println(\"hello\", 42) }\n```\n\n```java\n@Override\nvoid sample() {}\n```\n\n```kotlin\n@Test\nfun sample() {}\n```", 100)
	then_noError(t, actualErr)
	gui, actualErr := gocui.NewGui(gocui.NewGuiOpts{OutputMode: gocui.OutputTrue})
	then_noError(t, actualErr)
	defer gui.Close()
	gui.FgColor = gocui.GetColor("#000000")
	gui.BgColor = gocui.GetColor("#FFFFFF")
	gui.SetManagerFunc(func(gui *gocui.Gui) error {
		width, height := gui.Size()
		view, err := gui.SetView(viewDetailName, 0, 0, width-1, height-1, 0)
		if err != nil && !isUnknownViewError(err) {
			return err
		}
		view.Title = "Grey syntax and diff backgrounds"
		view.Editable = false
		view.Clear()
		fmt.Fprint(view, renderReviewDiffFile(file, nil, width-4))
		for _, header := range []struct{ path, text string }{
			{"main.kt", "fun sample() {"},
			{"main.rb", "def sample"},
			{"main.java", "@Override"},
			{"annotations.kt", "@ExtendWith(SpringExtension::class)"},
			{"imports.kt", "@Import("},
		} {
			headerFile := reviewDiffFile{Path: header.path, Hunks: []reviewDiffHunk{{
				Header: "@@ -1,1 +1,1 @@",
				Lines: []reviewDiffLine{
					{Kind: reviewDiffDeletionLine, Text: header.text, LeftLine: 1},
					{Kind: reviewDiffAdditionLine, Text: header.text, RightLine: 1},
				},
			}}}
			fmt.Fprint(view, "\n", renderReviewDiffFile(headerFile, nil, width-4))
		}
		fmt.Fprint(view, "\nMarkdown code:\n", markdown)
		return nil
	})
	stop := make(chan struct{})
	defer close(stop)
	actualErrors := make(chan error, 1)
	go func() {
		actualErr := runManualVisualUpdateAfterMainLoop(gui, stop, func(*gocui.Gui) error { return nil })
		if actualErr == nil {
			actualErr = signalTmuxWaitToken(ready)
		}
		if actualErr == nil {
			actualErr = waitForTmuxToken(done)
		}
		actualErrors <- actualErr
		gui.Update(func(*gocui.Gui) error { return gocui.ErrQuit })
	}()
	actualErr = gui.MainLoop()
	if actualErr != nil && !isQuitError(actualErr) {
		t.Fatalf("expected no GUI error, actual %v", actualErr)
	}
	then_noError(t, <-actualErrors)
}
