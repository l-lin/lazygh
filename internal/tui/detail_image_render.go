package tui

import "github.com/jesseduffield/gocui"

func (program *Program) syncRenderedDetailImages(gui *gocui.Gui) {
	if program.detailImageManager == nil {
		return
	}
	if gui == nil {
		program.detailImageManager.Sync(nil)
		return
	}

	detailView, actualErr := gui.View(viewDetailName)
	if actualErr != nil || detailView == nil {
		program.detailImageManager.Sync(nil)
		return
	}

	program.detailImageManager.Sync(program.currentDetailDocument(detailView).images)
}
