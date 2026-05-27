package tui

type Msg interface {
	isMsg()
}

type MsgAppStarted struct{}

type MsgNextSideView struct{}

type MsgPreviousSideView struct{}

type MsgFocusPanelView struct {
	Number int
}

type MsgMoveSideSelection struct {
	Delta int
}

type MsgMoveSideSelectionToTop struct{}

type MsgMoveSideSelectionToBottom struct{}

type MsgOpenSearch struct {
	Query string
}

type MsgSearchDraftChanged struct {
	Query string
}

type MsgSearchEditorInputRequested struct {
	Intent lineEditorIntent
}

type MsgSubmitSearch struct{}

type MsgCancelSearch struct{}

type MsgCloseSearch struct{}

type MsgActionsPopupSearchInputRequested struct {
	Intent lineEditorIntent
}

type MsgOpenActionsPopup struct {
	ActionCount int
}

type MsgCloseActionsPopup struct{}

type MsgFocusActionsPopupSearch struct{}

type MsgFocusActionsPopupList struct{}

type MsgActionsPopupActionRequested struct {
	Action actionsPopupAction
}

type MsgMoveActionsPopupSelection struct {
	Delta int
}

type MsgActionsPopupPageRequested struct {
	Kind pageNavigationKind
}

type MsgActionsPopupPageResolved struct {
	Kind     pageNavigationKind
	PageSize int
}

type MsgActionsPopupViewportRequested struct {
	Placement viewportPlacement
}

type MsgMoveActionsPopupSelectionToTop struct{}

type MsgMoveActionsPopupSelectionToBottom struct{}

func (MsgAppStarted) isMsg()                        {}
func (MsgNextSideView) isMsg()                      {}
func (MsgPreviousSideView) isMsg()                  {}
func (MsgFocusPanelView) isMsg()                    {}
func (MsgMoveSideSelection) isMsg()                 {}
func (MsgMoveSideSelectionToTop) isMsg()            {}
func (MsgMoveSideSelectionToBottom) isMsg()         {}
func (MsgOpenSearch) isMsg()                        {}
func (MsgSearchDraftChanged) isMsg()                {}
func (MsgSearchEditorInputRequested) isMsg()        {}
func (MsgSubmitSearch) isMsg()                      {}
func (MsgCancelSearch) isMsg()                      {}
func (MsgCloseSearch) isMsg()                       {}
func (MsgActionsPopupSearchInputRequested) isMsg()  {}
func (MsgOpenActionsPopup) isMsg()                  {}
func (MsgCloseActionsPopup) isMsg()                 {}
func (MsgFocusActionsPopupSearch) isMsg()           {}
func (MsgFocusActionsPopupList) isMsg()             {}
func (MsgActionsPopupActionRequested) isMsg()       {}
func (MsgMoveActionsPopupSelection) isMsg()         {}
func (MsgActionsPopupPageRequested) isMsg()         {}
func (MsgActionsPopupPageResolved) isMsg()          {}
func (MsgActionsPopupViewportRequested) isMsg()     {}
func (MsgMoveActionsPopupSelectionToTop) isMsg()    {}
func (MsgMoveActionsPopupSelectionToBottom) isMsg() {}
