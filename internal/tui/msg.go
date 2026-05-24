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

type MsgSubmitSearch struct{}

type MsgCancelSearch struct{}

type MsgCloseSearch struct{}

type MsgOpenActionsPopup struct {
	ActionCount int
}

type MsgCloseActionsPopup struct{}

type MsgFocusActionsPopupSearch struct{}

type MsgFocusActionsPopupList struct{}

type MsgMoveActionsPopupSelection struct {
	Delta int
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
func (MsgSubmitSearch) isMsg()                      {}
func (MsgCancelSearch) isMsg()                      {}
func (MsgCloseSearch) isMsg()                       {}
func (MsgOpenActionsPopup) isMsg()                  {}
func (MsgCloseActionsPopup) isMsg()                 {}
func (MsgFocusActionsPopupSearch) isMsg()           {}
func (MsgFocusActionsPopupList) isMsg()             {}
func (MsgMoveActionsPopupSelection) isMsg()         {}
func (MsgMoveActionsPopupSelectionToTop) isMsg()    {}
func (MsgMoveActionsPopupSelectionToBottom) isMsg() {}
