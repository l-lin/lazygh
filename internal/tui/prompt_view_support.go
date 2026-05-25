package tui

import (
	"fmt"
	"unicode/utf8"

	"github.com/jesseduffield/gocui"

	"github.com/l-lin/lazygh/internal/theme"
)

func (program *Program) configureBottomPromptView(view *gocui.View, editor gocui.Editor, editable bool) {
	view.Title = ""
	view.Frame = false
	view.FrameRunes = nil
	view.FrameColor = gocui.GetColor(theme.ActiveBorderHex)
	view.TitleColor = gocui.GetColor(theme.ActiveTextHex)
	view.FgColor = gocui.GetColor(theme.ActiveTextHex)
	view.BgColor = gocuiColorOrDefault(theme.BackgroundHex)
	view.Wrap = false
	view.Highlight = false
	view.Editable = editable
	view.Editor = editor
}

func (program *Program) renderBottomPromptView(view *gocui.View, text string, cursorIndex int) {
	if view == nil {
		return
	}

	view.Clear()
	prompt := bottomPromptPrefix + text
	fmt.Fprint(view, prompt)
	program.setInputCursor(view, prompt, cursorIndex+utf8.RuneCountInString(bottomPromptPrefix))
}

func (program *Program) setInputCursor(view *gocui.View, value string, cursorIndex int) {
	if view == nil {
		return
	}

	innerWidth := max(view.InnerWidth(), 1)

	valueWidth := utf8.RuneCountInString(value)
	if cursorIndex < 0 {
		cursorIndex = 0
	}
	if cursorIndex > valueWidth {
		cursorIndex = valueWidth
	}

	originX := 0
	if cursorIndex >= innerWidth {
		originX = cursorIndex - innerWidth + 1
	}
	cursorX := max(cursorIndex-originX, 0)
	if cursorX >= innerWidth {
		cursorX = innerWidth - 1
	}

	view.SetOrigin(originX, 0)
	view.SetCursor(cursorX, 0)
}
