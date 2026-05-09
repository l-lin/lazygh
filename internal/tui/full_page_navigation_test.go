package tui

import (
	"testing"

	"github.com/jesseduffield/gocui"
	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestKeybindingSpecs_GivenProgram_WhenListingFullPageBindings_ThenReadOnlyViewsSupportControlFControlBAndPageKeys(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := subject.keybindingSpecs()

	for _, viewName := range []string{viewUserName, viewPullRequestsName, viewDetailName} {
		then_bindingExists(t, actual, keybindingSpec{viewName: viewName, key: gocui.KeyCtrlF, handler: subject.fullPageDown})
		then_bindingExists(t, actual, keybindingSpec{viewName: viewName, key: gocui.KeyPgdn, handler: subject.fullPageDown})
		then_bindingExists(t, actual, keybindingSpec{viewName: viewName, key: gocui.KeyCtrlB, handler: subject.fullPageUp})
		then_bindingExists(t, actual, keybindingSpec{viewName: viewName, key: gocui.KeyPgup, handler: subject.fullPageUp})
	}

	then_bindingExists(t, actual, keybindingSpec{viewName: viewActionsPopupName, key: gocui.KeyCtrlF, handler: subject.fullPageActionsPopupDown})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewActionsPopupName, key: gocui.KeyPgdn, handler: subject.fullPageActionsPopupDown})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewActionsPopupName, key: gocui.KeyCtrlB, handler: subject.fullPageActionsPopupUp})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewActionsPopupName, key: gocui.KeyPgup, handler: subject.fullPageActionsPopupUp})

	then_bindingExists(t, actual, keybindingSpec{viewName: viewHelpName, key: gocui.KeyCtrlF, handler: subject.fullPageHelpDown})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewHelpName, key: gocui.KeyPgdn, handler: subject.fullPageHelpDown})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewHelpName, key: gocui.KeyCtrlB, handler: subject.fullPageHelpUp})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewHelpName, key: gocui.KeyPgup, handler: subject.fullPageHelpUp})
}

func TestKeybindingSpecs_GivenProgram_WhenListingFullPageBindings_ThenTextInputsDoNotCaptureControlFOrControlB(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := subject.keybindingSpecs()

	for _, viewName := range []string{viewSearchName, viewActionsPopupSearchName, viewModalEditorName} {
		then_bindingDoesNotExist(t, actual, viewName, gocui.KeyCtrlF)
		then_bindingDoesNotExist(t, actual, viewName, gocui.KeyCtrlB)
		then_bindingDoesNotExist(t, actual, viewName, gocui.KeyPgdn)
		then_bindingDoesNotExist(t, actual, viewName, gocui.KeyPgup)
	}
}

func TestFullPageNavigation_GivenUserViewSelection_WhenPressingControlFAndControlB_ThenItMovesAFullPageAndRecenters(t *testing.T) {
	subject := NewProgramWithModel(NewModel(SeedData{Users: given_manyItems("user", 60)}))
	gui := given_headlessGuiWithSize(t, 120, 12)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	userView, actualErr := gui.View(viewUserName)
	then_noError(t, actualErr)

	initialIndex := userView.InnerHeight() + 1
	for range initialIndex {
		subject.model.MoveSelectionDown()
	}
	actualErr = subject.layout(gui)
	then_noError(t, actualErr)
	userView, actualErr = gui.View(viewUserName)
	then_noError(t, actualErr)
	step := fullPageDelta(userView.InnerHeight())
	lineCount := len(subject.model.VisibleUsers())
	expectedDownIndex := clampIndex(initialIndex+step, lineCount)

	fullPageDownHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewUserName, gocui.KeyCtrlF)
	actualErr = fullPageDownHandler(gui, userView)
	then_noError(t, actualErr)
	actualErr = subject.layout(gui)
	then_noError(t, actualErr)
	userView, actualErr = gui.View(viewUserName)
	then_noError(t, actualErr)
	then_listViewIsCenteredOnSelection(t, userView, expectedDownIndex, lineCount)

	fullPageUpHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewUserName, gocui.KeyCtrlB)
	actualErr = fullPageUpHandler(gui, userView)
	then_noError(t, actualErr)
	actualErr = subject.layout(gui)
	then_noError(t, actualErr)
	userView, actualErr = gui.View(viewUserName)
	then_noError(t, actualErr)
	then_listViewIsCenteredOnSelection(t, userView, initialIndex, lineCount)
}

func TestFullPageNavigation_GivenReviewFilesViewSelection_WhenPressingPageDownAndPageUp_ThenItMovesAFullPageAndRecenters(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/widgets#42": given_reviewSessionPullRequestDiff(),
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGuiWithSize(t, 120, 12)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingReviewMode(t, gui, subject)
	then_noError(t, actualErr)
	filesView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)

	selectableRows, ok := subject.reviewSessionSelectableRows()
	if !ok || len(selectableRows) < 2 {
		t.Fatalf("expected selectable review rows, actual %v", selectableRows)
	}
	initialRow := subject.reviewSession.selectedFileTreeRow
	step := fullPageDelta(filesView.InnerHeight())
	expectedDownRow := adjustVisibleSelection(initialRow, selectableRows, step)
	expectedUpRow := adjustVisibleSelection(expectedDownRow, selectableRows, -step)
	lineCount := len(subject.reviewSessionFiles())
	if tree, _, ok := subject.reviewSessionCurrentTree(); ok {
		lineCount = len(tree.Rows)
	}

	fullPageDownHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewPullRequestsName, gocui.KeyPgdn)
	actualErr = fullPageDownHandler(gui, filesView)
	then_noError(t, actualErr)
	filesView, actualErr = gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	if subject.reviewSession.selectedFileTreeRow != expectedDownRow {
		t.Fatalf("expected selected review row %d, actual %d", expectedDownRow, subject.reviewSession.selectedFileTreeRow)
	}
	then_listViewIsCenteredOnSelection(t, filesView, expectedDownRow, lineCount)

	fullPageUpHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewPullRequestsName, gocui.KeyPgup)
	actualErr = fullPageUpHandler(gui, filesView)
	then_noError(t, actualErr)
	filesView, actualErr = gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	if subject.reviewSession.selectedFileTreeRow != expectedUpRow {
		t.Fatalf("expected selected review row %d after paging back up, actual %d", expectedUpRow, subject.reviewSession.selectedFileTreeRow)
	}
	then_listViewIsCenteredOnSelection(t, filesView, expectedUpRow, lineCount)
}

func TestFullPageNavigation_GivenDetailCursor_WhenPressingControlFAndControlB_ThenItMovesAFullPageAndRecenters(t *testing.T) {
	model := NewModel(SeedData{Users: []Item{{Title: "user-1", Detail: given_multilineDetail(60)}}})
	model.OpenDetail()
	subject := NewProgramWithModel(model)
	gui := given_headlessGuiWithSize(t, 120, 12)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)

	initialLine := detailView.InnerHeight() + 1
	for range initialLine {
		actualErr = subject.moveSelectionDown(gui, detailView)
		then_noError(t, actualErr)
	}
	detailView, actualErr = gui.View(viewDetailName)
	then_noError(t, actualErr)
	step := fullPageDelta(detailView.InnerHeight())
	expectedDownLine := clampIndex(initialLine+step, 60)

	fullPageDownHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, gocui.KeyCtrlF)
	actualErr = fullPageDownHandler(gui, detailView)
	then_noError(t, actualErr)
	detailView, actualErr = gui.View(viewDetailName)
	then_noError(t, actualErr)
	then_detailViewIsCenteredOnCursor(t, detailView, expectedDownLine, 60)

	fullPageUpHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, gocui.KeyCtrlB)
	actualErr = fullPageUpHandler(gui, detailView)
	then_noError(t, actualErr)
	detailView, actualErr = gui.View(viewDetailName)
	then_noError(t, actualErr)
	then_detailViewIsCenteredOnCursor(t, detailView, initialLine, 60)
}

func TestFullPageNavigation_GivenHelpView_WhenPressingPageDownAndPageUp_ThenItScrollsAFullPage(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	gui := given_headlessGuiWithSize(t, 80, 10)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.toggleHelp(gui, nil)
	then_noError(t, actualErr)
	helpView, actualErr := gui.View(viewHelpName)
	then_noError(t, actualErr)

	lineCount := len(helpView.BufferLines())
	step := fullPageDelta(helpView.InnerHeight())
	expectedDownOrigin := clampInt(step, 0, maxInt(0, lineCount-helpView.InnerHeight()))

	fullPageDownHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewHelpName, gocui.KeyPgdn)
	actualErr = fullPageDownHandler(gui, helpView)
	then_noError(t, actualErr)
	helpView, actualErr = gui.View(viewHelpName)
	then_noError(t, actualErr)
	then_viewOriginYIs(t, helpView, expectedDownOrigin)

	fullPageUpHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewHelpName, gocui.KeyPgup)
	actualErr = fullPageUpHandler(gui, helpView)
	then_noError(t, actualErr)
	helpView, actualErr = gui.View(viewHelpName)
	then_noError(t, actualErr)
	then_viewOriginYIs(t, helpView, 0)
}

func TestFullPageNavigation_GivenActionsPopupSelection_WhenPressingPageDownAndPageUp_ThenItMovesAFullPageAndRecenters(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	gui := given_headlessGuiWithSize(t, 120, 12)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)

	fullPageDownHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewActionsPopupName, gocui.KeyPgdn)
	actualErr = fullPageDownHandler(gui, popupView)
	then_noError(t, actualErr)
	popupView, actualErr = gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	then_listViewIsCenteredOnSelection(t, popupView, subject.currentActionsPopupSelectedRenderedLine(), subject.currentActionsPopupRenderedLineCount())

	fullPageUpHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewActionsPopupName, gocui.KeyPgup)
	actualErr = fullPageUpHandler(gui, popupView)
	then_noError(t, actualErr)
	popupView, actualErr = gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	then_listViewIsCenteredOnSelection(t, popupView, subject.currentActionsPopupSelectedRenderedLine(), subject.currentActionsPopupRenderedLineCount())
}

func TestSearchPrompt_GivenControlBAndControlF_WhenEditing_ThenTheyMoveTheCursor(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openSearch(gui, nil)
	then_noError(t, actualErr)
	searchView, actualErr := gui.View(viewSearchName)
	then_noError(t, actualErr)

	for _, character := range []rune{'a', 'b', 'c'} {
		actualHandled := subject.editSearch(searchView, 0, character, gocui.ModNone)
		if !actualHandled {
			t.Fatalf("expected %q to be handled", string(character))
		}
	}
	actual := subject.keybindingSpecs()
	then_bindingDoesNotExist(t, actual, viewSearchName, gocui.KeyCtrlB)
	then_bindingDoesNotExist(t, actual, viewSearchName, gocui.KeyCtrlF)

	actualHandled := subject.editSearch(searchView, gocui.KeyCtrlB, 0, gocui.ModNone)
	if !actualHandled {
		t.Fatal("expected ctrl-b to be handled by the search editor")
	}
	if subject.searchEditor.Cursor() != 2 {
		t.Fatalf("expected search cursor 2 after ctrl-b, actual %d", subject.searchEditor.Cursor())
	}

	actualHandled = subject.editSearch(searchView, gocui.KeyCtrlF, 0, gocui.ModNone)
	if !actualHandled {
		t.Fatal("expected ctrl-f to be handled by the search editor")
	}
	if subject.searchEditor.Cursor() != 3 {
		t.Fatalf("expected search cursor 3 after ctrl-f, actual %d", subject.searchEditor.Cursor())
	}
}

func TestActionsPopupSearch_GivenControlBAndControlF_WhenEditing_ThenTheyMoveTheCursor(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.focusActionsPopupSearch(gui, nil)
	then_noError(t, actualErr)
	searchView, actualErr := gui.View(viewActionsPopupSearchName)
	then_noError(t, actualErr)

	for _, character := range []rune{'a', 'b', 'c'} {
		actualHandled := subject.editActionsPopupSearch(searchView, 0, character, gocui.ModNone)
		if !actualHandled {
			t.Fatalf("expected %q to be handled", string(character))
		}
	}
	actual := subject.keybindingSpecs()
	then_bindingDoesNotExist(t, actual, viewActionsPopupSearchName, gocui.KeyCtrlB)
	then_bindingDoesNotExist(t, actual, viewActionsPopupSearchName, gocui.KeyCtrlF)

	actualHandled := subject.editActionsPopupSearch(searchView, gocui.KeyCtrlB, 0, gocui.ModNone)
	if !actualHandled {
		t.Fatal("expected ctrl-b to be handled by the actions popup search editor")
	}
	if subject.actionsPopupSearchEditor.Cursor() != 2 {
		t.Fatalf("expected actions popup search cursor 2 after ctrl-b, actual %d", subject.actionsPopupSearchEditor.Cursor())
	}

	actualHandled = subject.editActionsPopupSearch(searchView, gocui.KeyCtrlF, 0, gocui.ModNone)
	if !actualHandled {
		t.Fatal("expected ctrl-f to be handled by the actions popup search editor")
	}
	if subject.actionsPopupSearchEditor.Cursor() != 3 {
		t.Fatalf("expected actions popup search cursor 3 after ctrl-f, actual %d", subject.actionsPopupSearchEditor.Cursor())
	}
}

func TestModalEditor_GivenControlBAndControlF_WhenEditing_ThenTheyMoveTheCursor(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openModalEditor(gui, "Edit", "abc", func(string) error { return nil })
	then_noError(t, actualErr)
	modalView, actualErr := gui.View(viewModalEditorName)
	then_noError(t, actualErr)

	actual := subject.keybindingSpecs()
	then_bindingDoesNotExist(t, actual, viewModalEditorName, gocui.KeyCtrlB)
	then_bindingDoesNotExist(t, actual, viewModalEditorName, gocui.KeyCtrlF)

	actualHandled := subject.editModalEditor(modalView, gocui.KeyCtrlB, 0, gocui.ModNone)
	if !actualHandled {
		t.Fatal("expected ctrl-b to be handled by the modal editor")
	}
	actualColumn, actualRow := subject.modalEditor.CursorXY()
	if actualColumn != 2 || actualRow != 0 {
		t.Fatalf("expected modal editor cursor 2,0 after ctrl-b, actual %d,%d", actualColumn, actualRow)
	}

	actualHandled = subject.editModalEditor(modalView, gocui.KeyCtrlF, 0, gocui.ModNone)
	if !actualHandled {
		t.Fatal("expected ctrl-f to be handled by the modal editor")
	}
	actualColumn, actualRow = subject.modalEditor.CursorXY()
	if actualColumn != 3 || actualRow != 0 {
		t.Fatalf("expected modal editor cursor 3,0 after ctrl-f, actual %d,%d", actualColumn, actualRow)
	}
}

func then_viewOriginYIs(t *testing.T, view *gocui.View, expectedOriginY int) {
	t.Helper()

	_, actualOriginY := view.Origin()
	if actualOriginY != expectedOriginY {
		t.Fatalf("expected origin y %d, actual %d", expectedOriginY, actualOriginY)
	}
}
