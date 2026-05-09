package tui

import (
	"strings"

	"github.com/jesseduffield/gocui"

	"github.com/l-lin/lazygh/internal/theme"
)

func configureFramedOverlayView(view *gocui.View, title string, footer string) {
	if view == nil {
		return
	}

	view.Title = strings.TrimSpace(title)
	view.Footer = strings.TrimSpace(footer)
	view.Frame = true
	view.FrameRunes = roundFrameRunes
	view.FrameColor = gocui.GetColor(theme.ActiveBorderHex)
	view.TitleColor = gocui.GetColor(theme.ActiveTextHex)
	view.FgColor = gocui.GetColor(theme.ActiveTextHex)
	view.BgColor = gocuiColorOrDefault(theme.BackgroundHex)
}
