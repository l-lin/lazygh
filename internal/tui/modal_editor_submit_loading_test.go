package tui

import (
	"testing"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func TestModalEditorSubmitRequests_GivenMutationRequests_WhenReadingTheirLoadingSurface_ThenTheyExposeTheExpectedGHCommands(t *testing.T) {
	tests := []struct {
		name            string
		request         modalEditorSubmitRequest
		expectedCommand string
		expectedAsync   bool
	}{
		{
			name:            "pull request comment",
			request:         pullRequestCommentSubmitRequest{target: pullRequestCommentTarget{repository: "acme/widgets", number: 42}, body: "Ship it"},
			expectedCommand: "gh pr comment 42 -R acme/widgets --body-file -",
			expectedAsync:   true,
		},
		{
			name:            "pull request review comment",
			request:         pullRequestReviewCommentSubmitRequest{target: pullRequestActionTarget{repository: "acme/widgets", number: 42}, body: "Needs a note"},
			expectedCommand: "gh pr review 42 -R acme/widgets --comment --body-file -",
			expectedAsync:   true,
		},
		{
			name:            "pull request request changes",
			request:         pullRequestRequestChangesSubmitRequest{target: pullRequestActionTarget{repository: "acme/widgets", number: 42}, body: "Please fix"},
			expectedCommand: "gh pr review 42 -R acme/widgets --request-changes --body-file -",
			expectedAsync:   true,
		},
		{
			name:            "pull request title edit",
			request:         pullRequestTitleEditSubmitRequest{target: pullRequestActionTarget{repository: "acme/widgets", number: 42}, title: "Rename me"},
			expectedCommand: "gh pr edit 42 -R acme/widgets --title Rename me",
			expectedAsync:   true,
		},
		{
			name:            "pull request description edit",
			request:         pullRequestDescriptionEditSubmitRequest{target: pullRequestActionTarget{repository: "acme/widgets", number: 42}, body: "Updated body"},
			expectedCommand: "gh pr edit 42 -R acme/widgets --body-file -",
			expectedAsync:   true,
		},
		{
			name:            "pull request comment update",
			request:         pullRequestCommentUpdateSubmitRequest{target: pullRequestCommentEditActionTarget{commentID: "comment-1"}, body: "Updated comment"},
			expectedCommand: "gh api graphql",
			expectedAsync:   true,
		},
		{
			name:            "inline comment update",
			request:         inlineCommentUpdateSubmitRequest{target: pullRequestReviewCommentActionTarget{commentID: "comment-2"}, body: "Updated inline comment"},
			expectedCommand: "gh api graphql",
			expectedAsync:   true,
		},
		{
			name:            "inline comment reply",
			request:         inlineCommentReplySubmitRequest{target: pullRequestReviewThreadReplyTarget{repository: "acme/widgets", number: 42, threadID: "thread-1"}, body: "Reply body"},
			expectedCommand: "gh api graphql",
			expectedAsync:   true,
		},
		{
			name:            "inline comment pending review lookup",
			request:         reviewInlineCommentSubmitRequest{target: pullRequestInlineCommentTarget{repository: "acme/widgets", number: 42}, body: "Inline feedback"},
			expectedCommand: "gh api graphql",
			expectedAsync:   true,
		},
		{
			name:            "prepared inline comment submit",
			request:         preparedReviewInlineCommentSubmitRequest{target: pullRequestInlineCommentTarget{repository: "acme/widgets", number: 42, pendingReview: "PRR_pending"}, body: "Inline feedback"},
			expectedCommand: "gh api graphql",
			expectedAsync:   true,
		},
		{
			name:            "pending review submit",
			request:         pendingPullRequestReviewSubmitRequest{target: pendingPullRequestReviewTarget{repository: "acme/widgets", number: 42, pendingReviewID: "PRR_pending"}, event: githubdomain.PullRequestReviewEventComment, body: "Wrap up"},
			expectedCommand: "gh api graphql",
			expectedAsync:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actualCommand := test.request.statusCommand()
			if actualCommand != test.expectedCommand {
				t.Fatalf("expected command %q, actual %q", test.expectedCommand, actualCommand)
			}

			actualAsync := test.request.asyncRequested()
			if actualAsync != test.expectedAsync {
				t.Fatalf("expected async %v, actual %v", test.expectedAsync, actualAsync)
			}
		})
	}
}

func TestModalEditorSubmitRequests_GivenLocalRequests_WhenReadingTheirLoadingSurface_ThenTheyRemainSynchronousWithoutGHCommands(t *testing.T) {
	tests := []struct {
		name          string
		request       modalEditorSubmitRequest
		expectedAsync bool
	}{
		{
			name:          "open pull request by url",
			request:       openPullRequestByURLSubmitRequest{rawURL: "https://github.com/acme/widgets/pull/42"},
			expectedAsync: false,
		},
		{
			name:          "pull request custom search",
			request:       pullRequestCustomSearchSubmitRequest{criteria: "label:bug"},
			expectedAsync: false,
		},
		{
			name:          "inline comment submit with cached pending review",
			request:       reviewInlineCommentSubmitRequest{target: pullRequestInlineCommentTarget{repository: "acme/widgets", number: 42, pendingReview: "PRR_pending"}, body: "Inline feedback"},
			expectedAsync: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actualCommand := test.request.statusCommand()
			if actualCommand != "" {
				t.Fatalf("expected no command, actual %q", actualCommand)
			}

			actualAsync := test.request.asyncRequested()
			if actualAsync != test.expectedAsync {
				t.Fatalf("expected async %v, actual %v", test.expectedAsync, actualAsync)
			}
		})
	}
}
