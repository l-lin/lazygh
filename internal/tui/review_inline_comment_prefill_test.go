package tui

import (
	"testing"
)

func TestPullRequestInlineCommentSelectionFromRenderedRows_GivenALinewiseSelectionOnAWrappedReviewDiffContinuationRow_WhenBuildingTheInitialSuggestionBody_ThenItUsesTheSelectedDiffLine(t *testing.T) {
	given_file := reviewDiffFile{
		Path:       "src/main/java/com/acme/Imports.java",
		Additions:  3,
		Deletions:  0,
		ChangeType: reviewDiffChangeTypeModified,
		Hunks: []reviewDiffHunk{{
			Header: "@@ -10,0 +10,3 @@",
			Lines: []reviewDiffLine{
				{Kind: reviewDiffAdditionLine, Text: "import com.doctolib.doctoboot.interservices.client.v2.SomeRidiculouslyLongPreviousImportThatWraps;", RightLine: 10, Side: reviewDiffLineSideRight},
				{Kind: reviewDiffAdditionLine, Text: "import com.doctolib.doctoboot.interservices.client.v2.JwtCredentialManagerFactoryV2;", RightLine: 11, Side: reviewDiffLineSideRight},
				{Kind: reviewDiffAdditionLine, Text: "import com.doctolib.doctoboot.rest.InterServiceJwtInterceptor;", RightLine: 12, Side: reviewDiffLineSideRight},
			},
		}},
	}
	given_renderedRows := buildReviewDiffRenderedRows(given_file, nil, 72)
	given_document := newReviewDiffDetailDocumentWithWordWrap(given_renderedRows, 42, true)
	selectedLineIndex, _ := given_detailDocumentLineContaining(t, given_document, "JwtCredentialManagerFactoryV2")
	given_wrappedContinuationColumn := given_document.wrapWidthForLine(selectedLineIndex) + 1
	given_state := newDetailViewState()
	given_state.mode = detailLineVisualMode
	given_state.visualAnchor = detailPosition{line: selectedLineIndex, column: given_wrappedContinuationColumn}
	given_state.cursor = detailPosition{line: selectedLineIndex, column: given_wrappedContinuationColumn}
	given_state.sync(given_document, 20)

	actual, actualErr := pullRequestInlineCommentSelectionFromRenderedRows("acme/widgets", 42, "PRR_pending", false, given_renderedRows, given_document, given_state)

	then_noError(t, actualErr)
	expected := "```suggestion\nimport com.doctolib.doctoboot.interservices.client.v2.JwtCredentialManagerFactoryV2;\n```"
	if actual.initialBody != expected {
		t.Fatalf("expected inline comment draft %q, actual %q", expected, actual.initialBody)
	}
	if actual.target.threadTarget.Line != 11 {
		t.Fatalf("expected selected target line %d, actual %d", 11, actual.target.threadTarget.Line)
	}
}
