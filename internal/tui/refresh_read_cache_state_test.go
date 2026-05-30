package tui

import "testing"

func TestRefreshReadCacheState_GivenPopupMemoization_WhenCachingActionsAndVisibleLines_ThenItReturnsUpdatedCopiesWithoutMutatingTheOriginal(t *testing.T) {
	subject := refreshReadCacheState{enabled: true}
	givenActions := []actionsPopupAction{{id: "reply-to-inline-comment", title: "Reply to inline comment"}}
	givenVisibleLines := []actionsPopupVisibleLine{{text: "Reply to inline comment", actionIndex: 0, selectable: true}}

	actual := subject.withActionsPopupActions(givenActions).withActionsPopupVisibleLines(givenVisibleLines)

	if !actual.actionsPopupActionsKnown {
		t.Fatal("expected popup actions to be marked as memoized")
	}
	if !actual.actionsPopupVisibleKnown {
		t.Fatal("expected popup visible lines to be marked as memoized")
	}
	if actualCount := len(actual.actionsPopupActions); actualCount != 1 {
		t.Fatalf("expected one memoized action, actual %d", actualCount)
	}
	if actualCount := len(actual.actionsPopupVisibleLines); actualCount != 1 {
		t.Fatalf("expected one memoized visible line, actual %d", actualCount)
	}
	if subject.actionsPopupActionsKnown {
		t.Fatal("expected the original state to keep popup actions unknown")
	}
	if subject.actionsPopupVisibleKnown {
		t.Fatal("expected the original state to keep popup visible lines unknown")
	}
	if actualCount := len(subject.actionsPopupActions); actualCount != 0 {
		t.Fatalf("expected the original state to keep zero popup actions, actual %d", actualCount)
	}
	if actualCount := len(subject.actionsPopupVisibleLines); actualCount != 0 {
		t.Fatalf("expected the original state to keep zero popup visible lines, actual %d", actualCount)
	}
}

func TestRefreshReadCacheState_GivenPresenterSnapshots_WhenCachingResolverFooterPopupAndReviewModel_ThenItReturnsUpdatedCopiesWithoutMutatingTheOriginal(t *testing.T) {
	subject := refreshReadCacheState{enabled: true}
	givenReviewModel := reviewSessionReadModel{active: true}
	givenResolver := keybindingLabelResolver{}
	givenFooter := footerPresenter{modalEditorVisible: true}
	givenPopup := actionsPopupPresenter{searchQuery: "reply"}

	actual := subject.withReviewSessionReadModel(givenReviewModel).withKeybindingResolver(givenResolver).withFooterPresenter(givenFooter).withActionsPopupPresenter(givenPopup)

	if !actual.reviewSessionReadModelSet {
		t.Fatal("expected the review-session read model to be memoized")
	}
	if !actual.keybindingResolverSet {
		t.Fatal("expected the keybinding resolver to be memoized")
	}
	if !actual.footerPresenterSet {
		t.Fatal("expected the footer presenter to be memoized")
	}
	if !actual.actionsPopupPresenterSet {
		t.Fatal("expected the popup presenter to be memoized")
	}
	if !actual.reviewSessionReadModel.active {
		t.Fatal("expected the memoized review-session model to stay active")
	}
	if !actual.footerPresenter.modalEditorVisible {
		t.Fatal("expected the memoized footer presenter to keep the modal-editor visibility flag")
	}
	if actualQuery := actual.actionsPopupPresenter.searchQuery; actualQuery != "reply" {
		t.Fatalf("expected popup presenter query %q, actual %q", "reply", actualQuery)
	}
	if subject.reviewSessionReadModelSet {
		t.Fatal("expected the original state to keep the review-session model unset")
	}
	if subject.keybindingResolverSet {
		t.Fatal("expected the original state to keep the keybinding resolver unset")
	}
	if subject.footerPresenterSet {
		t.Fatal("expected the original state to keep the footer presenter unset")
	}
	if subject.actionsPopupPresenterSet {
		t.Fatal("expected the original state to keep the popup presenter unset")
	}
}
