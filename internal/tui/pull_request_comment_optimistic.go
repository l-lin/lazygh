package tui

import (
	"fmt"
	"strings"
	"time"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

const optimisticPullRequestMutationIDPrefix = "optimistic:"

func pullRequestMutationCacheKey(repository string, number int) string {
	trimmedRepository := strings.TrimSpace(repository)
	if trimmedRepository == "" || trimmedRepository == "-" || number <= 0 {
		return ""
	}
	return fmt.Sprintf("%s#%d", trimmedRepository, number)
}

func (program *Program) nextOptimisticPullRequestMutationID(kind string) string {
	if program == nil || program.optimisticMutationCoordinator == nil {
		return optimisticPullRequestMutationIDPrefix + strings.TrimSpace(kind)
	}
	updatedCoordinator, actualID := program.optimisticMutationCoordinator.nextOptimisticMutationID(kind)
	program.optimisticMutationCoordinator = &updatedCoordinator
	return actualID
}

func isOptimisticPullRequestMutationID(id string) bool {
	return strings.HasPrefix(strings.TrimSpace(id), optimisticPullRequestMutationIDPrefix)
}

func hasUsablePullRequestMutationID(id string) bool {
	trimmedID := strings.TrimSpace(id)
	return trimmedID != "" && !isOptimisticPullRequestMutationID(trimmedID)
}

func (program *Program) optimisticPullRequestCommentAuthor() *githubdomain.PullRequestCommentAuthor {
	login := program.currentConnectedUserLogin()
	if login == "" {
		return nil
	}
	return &githubdomain.PullRequestCommentAuthor{Login: login}
}

func (program *Program) newOptimisticPullRequestComment(body string, pending bool) githubdomain.PullRequestComment {
	state := ""
	if pending {
		state = "PENDING"
	}
	return githubdomain.PullRequestComment{
		ID:              program.nextOptimisticPullRequestMutationID("comment"),
		ViewerDidAuthor: true,
		Author:          program.optimisticPullRequestCommentAuthor(),
		Body:            body,
		CreatedAt:       time.Now().UTC().Format(time.RFC3339),
		State:           state,
	}
}

func (program *Program) optimisticallyAppendPullRequestComment(target pullRequestCommentTarget, body string) {
	comment := program.newOptimisticPullRequestComment(body, false)
	_ = program.mutatePullRequestDetailOptimistically(target.repository, target.number, func(detail *githubdomain.PullRequestDetail) bool {
		detail.Comments = append(detail.Comments, comment)
		return true
	})
}

func (program *Program) optimisticallyUpdatePullRequestComment(target pullRequestCommentEditActionTarget, body string) {
	_ = program.mutatePullRequestDetailOptimistically(target.repository, target.number, func(detail *githubdomain.PullRequestDetail) bool {
		return updatePullRequestCommentInPullRequestDetail(detail, target.commentID, body)
	})
}

func (program *Program) optimisticallyDeletePullRequestComment(target pullRequestCommentEditActionTarget) {
	_ = program.mutatePullRequestDetailOptimistically(target.repository, target.number, func(detail *githubdomain.PullRequestDetail) bool {
		return deletePullRequestCommentFromPullRequestDetail(detail, target.commentID)
	})
}

func (program *Program) optimisticallyAppendInlineCommentReply(target pullRequestReviewThreadReplyTarget, body string) {
	reply := program.newOptimisticPullRequestComment(body, strings.TrimSpace(target.pendingReview) != "")
	_ = program.mutatePullRequestDetailOptimistically(target.repository, target.number, func(detail *githubdomain.PullRequestDetail) bool {
		return appendReplyToPullRequestDetail(detail, target.threadID, reply)
	})
	_ = program.mutatePullRequestDiffOptimistically(target.repository, target.number, func(data *reviewDiffData) bool {
		return appendReplyToPullRequestDiff(data, target.threadID, reply)
	})
}

func (program *Program) optimisticallyAppendInlineComment(target pullRequestInlineCommentTarget, body string) {
	comment := program.newOptimisticPullRequestComment(body, true)
	threadID := program.nextOptimisticPullRequestMutationID("thread")
	threadTarget, ok := toDomainReviewThreadTarget(target.threadTarget)
	if !ok {
		return
	}
	thread := newOptimisticPullRequestReviewThread(threadTarget, comment, threadID)
	if target.updateDetail {
		_ = program.mutatePullRequestDetailOptimistically(target.repository, target.number, func(detail *githubdomain.PullRequestDetail) bool {
			return appendThreadToPullRequestDetail(detail, thread)
		})
	}
	_ = program.mutatePullRequestDiffOptimistically(target.repository, target.number, func(data *reviewDiffData) bool {
		return appendThreadToPullRequestDiff(data, buildReviewDiffThread(thread))
	})
}

func (program *Program) optimisticallyAddReaction(target pullRequestReactionActionTarget, content githubdomain.ReactionContent) {
	update := func(groups []githubdomain.ReactionGroup) []githubdomain.ReactionGroup {
		return optimisticReactionGroupsWithAddedReaction(groups, content)
	}
	_ = program.mutatePullRequestDetailOptimistically(target.repository, target.number, func(detail *githubdomain.PullRequestDetail) bool {
		return updateReactionGroupsInPullRequestDetail(detail, target.subjectID, update)
	})
	if target.invalidateDiff {
		_ = program.mutatePullRequestDiffOptimistically(target.repository, target.number, func(data *reviewDiffData) bool {
			return updateReactionGroupsInPullRequestDiff(data, target.subjectID, update)
		})
	}
}

func (program *Program) optimisticallyRemoveReaction(target pullRequestReactionRemovalTarget) {
	update := func(groups []githubdomain.ReactionGroup) []githubdomain.ReactionGroup {
		return optimisticReactionGroupsWithRemovedReaction(groups, target.content)
	}
	_ = program.mutatePullRequestDetailOptimistically(target.repository, target.number, func(detail *githubdomain.PullRequestDetail) bool {
		return updateReactionGroupsInPullRequestDetail(detail, target.subjectID, update)
	})
	if target.invalidateDiff {
		_ = program.mutatePullRequestDiffOptimistically(target.repository, target.number, func(data *reviewDiffData) bool {
			return updateReactionGroupsInPullRequestDiff(data, target.subjectID, update)
		})
	}
}

func (program *Program) optimisticallySetReviewThreadResolved(target pullRequestReviewThreadActionTarget, resolved bool) {
	_ = program.mutatePullRequestDetailOptimistically(target.repository, target.number, func(detail *githubdomain.PullRequestDetail) bool {
		return updateReviewThreadResolutionInPullRequestDetail(detail, target.threadID, resolved)
	})
	_ = program.mutatePullRequestDiffOptimistically(target.repository, target.number, func(data *reviewDiffData) bool {
		return updateReviewThreadResolutionInPullRequestDiff(data, target.threadID, resolved)
	})
}

func (program *Program) optimisticallyUpdateReviewComment(target pullRequestReviewCommentActionTarget, body string) {
	_ = program.mutatePullRequestDetailOptimistically(target.repository, target.number, func(detail *githubdomain.PullRequestDetail) bool {
		return updateReviewCommentInPullRequestDetail(detail, target.commentID, body)
	})
	_ = program.mutatePullRequestDiffOptimistically(target.repository, target.number, func(data *reviewDiffData) bool {
		return updateReviewCommentInPullRequestDiff(data, target.commentID, body)
	})
}

func (program *Program) optimisticallyDeleteReviewComment(target pullRequestReviewCommentActionTarget) {
	_ = program.mutatePullRequestDetailOptimistically(target.repository, target.number, func(detail *githubdomain.PullRequestDetail) bool {
		return deleteReviewCommentFromPullRequestDetail(detail, target.commentID)
	})
	_ = program.mutatePullRequestDiffOptimistically(target.repository, target.number, func(data *reviewDiffData) bool {
		return deleteReviewCommentFromPullRequestDiff(data, target.commentID)
	})
}

func newOptimisticPullRequestReviewThread(target githubdomain.PullRequestReviewThreadTarget, comment githubdomain.PullRequestComment, threadID string) githubdomain.PullRequestReviewThread {
	thread := githubdomain.PullRequestReviewThread{
		ID:            strings.TrimSpace(threadID),
		Path:          strings.TrimSpace(target.Path),
		Line:          target.Line,
		StartLine:     target.StartLine,
		DiffSide:      strings.TrimSpace(target.Side),
		StartDiffSide: strings.TrimSpace(target.StartSide),
		Comments:      []githubdomain.PullRequestComment{comment},
	}
	if strings.EqualFold(thread.DiffSide, string(reviewDiffLineSideLeft)) {
		thread.OriginalLine = thread.Line
	}
	if thread.StartLine > 0 && strings.EqualFold(thread.StartDiffSide, string(reviewDiffLineSideLeft)) {
		thread.OriginalStartLine = thread.StartLine
	}
	return thread
}

func appendReplyToPullRequestDetail(detail *githubdomain.PullRequestDetail, threadID string, reply githubdomain.PullRequestComment) bool {
	if detail == nil || !hasUsablePullRequestMutationID(threadID) {
		return false
	}

	trimmedThreadID := strings.TrimSpace(threadID)
	for threadIndex := range detail.InlineCommentThreads {
		if strings.TrimSpace(detail.InlineCommentThreads[threadIndex].ID) != trimmedThreadID {
			continue
		}
		detail.InlineCommentThreads[threadIndex].Comments = append(detail.InlineCommentThreads[threadIndex].Comments, reply)
		return true
	}
	return false
}

func appendReplyToPullRequestDiff(data *reviewDiffData, threadID string, reply githubdomain.PullRequestComment) bool {
	if data == nil || !hasUsablePullRequestMutationID(threadID) {
		return false
	}

	trimmedThreadID := strings.TrimSpace(threadID)
	for fileIndex := range data.Files {
		for threadIndex := range data.Files[fileIndex].Threads {
			if strings.TrimSpace(data.Files[fileIndex].Threads[threadIndex].ID) != trimmedThreadID {
				continue
			}
			data.Files[fileIndex].Threads[threadIndex].Comments = append(data.Files[fileIndex].Threads[threadIndex].Comments, reply)
			return true
		}
	}
	return false
}

func appendThreadToPullRequestDetail(detail *githubdomain.PullRequestDetail, thread githubdomain.PullRequestReviewThread) bool {
	if detail == nil || strings.TrimSpace(thread.Path) == "" {
		return false
	}

	detail.InlineCommentThreads = append(detail.InlineCommentThreads, thread)
	return true
}

func appendThreadToPullRequestDiff(data *reviewDiffData, thread reviewDiffThread) bool {
	if data == nil || strings.TrimSpace(thread.Path) == "" {
		return false
	}

	trimmedPath := strings.TrimSpace(thread.Path)
	for fileIndex := range data.Files {
		if strings.TrimSpace(data.Files[fileIndex].Path) != trimmedPath {
			continue
		}
		data.Files[fileIndex].Threads = append(data.Files[fileIndex].Threads, thread)
		return true
	}
	return false
}

func updateReactionGroupsInPullRequestDetail(detail *githubdomain.PullRequestDetail, subjectID string, update func([]githubdomain.ReactionGroup) []githubdomain.ReactionGroup) bool {
	trimmedSubjectID := strings.TrimSpace(subjectID)
	if detail == nil || trimmedSubjectID == "" || update == nil {
		return false
	}

	updated := false
	if strings.TrimSpace(detail.ID) == trimmedSubjectID {
		detail.ReactionGroups = update(detail.ReactionGroups)
		updated = true
	}
	for index := range detail.Comments {
		if strings.TrimSpace(detail.Comments[index].ID) != trimmedSubjectID {
			continue
		}
		detail.Comments[index].ReactionGroups = update(detail.Comments[index].ReactionGroups)
		updated = true
	}
	for index := range detail.InlineComments {
		if strings.TrimSpace(detail.InlineComments[index].ID) != trimmedSubjectID {
			continue
		}
		detail.InlineComments[index].ReactionGroups = update(detail.InlineComments[index].ReactionGroups)
		updated = true
	}
	for threadIndex := range detail.InlineCommentThreads {
		for commentIndex := range detail.InlineCommentThreads[threadIndex].Comments {
			if strings.TrimSpace(detail.InlineCommentThreads[threadIndex].Comments[commentIndex].ID) != trimmedSubjectID {
				continue
			}
			detail.InlineCommentThreads[threadIndex].Comments[commentIndex].ReactionGroups = update(detail.InlineCommentThreads[threadIndex].Comments[commentIndex].ReactionGroups)
			updated = true
		}
	}
	return updated
}

func updateReactionGroupsInPullRequestDiff(data *reviewDiffData, subjectID string, update func([]githubdomain.ReactionGroup) []githubdomain.ReactionGroup) bool {
	trimmedSubjectID := strings.TrimSpace(subjectID)
	if data == nil || trimmedSubjectID == "" || update == nil {
		return false
	}

	updated := false
	for fileIndex := range data.Files {
		for threadIndex := range data.Files[fileIndex].Threads {
			for commentIndex := range data.Files[fileIndex].Threads[threadIndex].Comments {
				if strings.TrimSpace(data.Files[fileIndex].Threads[threadIndex].Comments[commentIndex].ID) != trimmedSubjectID {
					continue
				}
				data.Files[fileIndex].Threads[threadIndex].Comments[commentIndex].ReactionGroups = update(data.Files[fileIndex].Threads[threadIndex].Comments[commentIndex].ReactionGroups)
				updated = true
			}
		}
	}
	return updated
}

func optimisticReactionGroupsWithAddedReaction(groups []githubdomain.ReactionGroup, content githubdomain.ReactionContent) []githubdomain.ReactionGroup {
	updatedGroups := append([]githubdomain.ReactionGroup(nil), groups...)
	trimmedContent := strings.TrimSpace(string(content))
	for index := range updatedGroups {
		if strings.TrimSpace(string(updatedGroups[index].Content)) != trimmedContent {
			continue
		}
		updatedGroups[index].TotalCount++
		if updatedGroups[index].TotalCount < 1 {
			updatedGroups[index].TotalCount = 1
		}
		updatedGroups[index].ViewerHasReacted = true
		return updatedGroups
	}
	return append(updatedGroups, githubdomain.ReactionGroup{Content: content, TotalCount: 1, ViewerHasReacted: true})
}

func optimisticReactionGroupsWithRemovedReaction(groups []githubdomain.ReactionGroup, content githubdomain.ReactionContent) []githubdomain.ReactionGroup {
	updatedGroups := append([]githubdomain.ReactionGroup(nil), groups...)
	trimmedContent := strings.TrimSpace(string(content))
	filteredGroups := updatedGroups[:0]
	for _, group := range updatedGroups {
		if strings.TrimSpace(string(group.Content)) != trimmedContent {
			filteredGroups = append(filteredGroups, group)
			continue
		}
		if group.TotalCount <= 1 {
			continue
		}
		group.TotalCount--
		group.ViewerHasReacted = false
		filteredGroups = append(filteredGroups, group)
	}
	return append([]githubdomain.ReactionGroup(nil), filteredGroups...)
}

func updateReviewThreadResolutionInPullRequestDetail(detail *githubdomain.PullRequestDetail, threadID string, resolved bool) bool {
	trimmedThreadID := strings.TrimSpace(threadID)
	if detail == nil || !hasUsablePullRequestMutationID(trimmedThreadID) {
		return false
	}

	updated := false
	for index := range detail.InlineCommentThreads {
		if strings.TrimSpace(detail.InlineCommentThreads[index].ID) != trimmedThreadID {
			continue
		}
		detail.InlineCommentThreads[index].IsResolved = resolved
		updated = true
	}
	return updated
}

func updateReviewThreadResolutionInPullRequestDiff(data *reviewDiffData, threadID string, resolved bool) bool {
	trimmedThreadID := strings.TrimSpace(threadID)
	if data == nil || !hasUsablePullRequestMutationID(trimmedThreadID) {
		return false
	}

	updated := false
	for fileIndex := range data.Files {
		for threadIndex := range data.Files[fileIndex].Threads {
			if strings.TrimSpace(data.Files[fileIndex].Threads[threadIndex].ID) != trimmedThreadID {
				continue
			}
			data.Files[fileIndex].Threads[threadIndex].IsResolved = resolved
			updated = true
		}
	}
	return updated
}

func updatePullRequestCommentInPullRequestDetail(detail *githubdomain.PullRequestDetail, commentID string, body string) bool {
	trimmedCommentID := strings.TrimSpace(commentID)
	if detail == nil || !hasUsablePullRequestMutationID(trimmedCommentID) {
		return false
	}

	updated := false
	for commentIndex := range detail.Comments {
		if strings.TrimSpace(detail.Comments[commentIndex].ID) != trimmedCommentID {
			continue
		}
		detail.Comments[commentIndex].Body = body
		detail.Comments[commentIndex].BodyHTML = ""
		updated = true
	}
	return updated
}

func updateReviewCommentInPullRequestDetail(detail *githubdomain.PullRequestDetail, commentID string, body string) bool {
	trimmedCommentID := strings.TrimSpace(commentID)
	if detail == nil || !hasUsablePullRequestMutationID(trimmedCommentID) {
		return false
	}

	updated := false
	for threadIndex := range detail.InlineCommentThreads {
		for commentIndex := range detail.InlineCommentThreads[threadIndex].Comments {
			if strings.TrimSpace(detail.InlineCommentThreads[threadIndex].Comments[commentIndex].ID) != trimmedCommentID {
				continue
			}
			detail.InlineCommentThreads[threadIndex].Comments[commentIndex].Body = body
			detail.InlineCommentThreads[threadIndex].Comments[commentIndex].BodyHTML = ""
			updated = true
		}
	}
	for commentIndex := range detail.InlineComments {
		if strings.TrimSpace(detail.InlineComments[commentIndex].ID) != trimmedCommentID {
			continue
		}
		detail.InlineComments[commentIndex].Body = body
		detail.InlineComments[commentIndex].BodyHTML = ""
		updated = true
	}
	return updated
}

func updateReviewCommentInPullRequestDiff(data *reviewDiffData, commentID string, body string) bool {
	trimmedCommentID := strings.TrimSpace(commentID)
	if data == nil || !hasUsablePullRequestMutationID(trimmedCommentID) {
		return false
	}

	updated := false
	for fileIndex := range data.Files {
		for threadIndex := range data.Files[fileIndex].Threads {
			for commentIndex := range data.Files[fileIndex].Threads[threadIndex].Comments {
				if strings.TrimSpace(data.Files[fileIndex].Threads[threadIndex].Comments[commentIndex].ID) != trimmedCommentID {
					continue
				}
				data.Files[fileIndex].Threads[threadIndex].Comments[commentIndex].Body = body
				data.Files[fileIndex].Threads[threadIndex].Comments[commentIndex].BodyHTML = ""
				updated = true
			}
		}
	}
	return updated
}

func deletePullRequestCommentFromPullRequestDetail(detail *githubdomain.PullRequestDetail, commentID string) bool {
	trimmedCommentID := strings.TrimSpace(commentID)
	if detail == nil || !hasUsablePullRequestMutationID(trimmedCommentID) {
		return false
	}

	updated := false
	filteredComments := detail.Comments[:0]
	for _, comment := range detail.Comments {
		if strings.TrimSpace(comment.ID) == trimmedCommentID {
			updated = true
			continue
		}
		filteredComments = append(filteredComments, comment)
	}
	detail.Comments = filteredComments
	return updated
}

func deleteReviewCommentFromPullRequestDetail(detail *githubdomain.PullRequestDetail, commentID string) bool {
	trimmedCommentID := strings.TrimSpace(commentID)
	if detail == nil || !hasUsablePullRequestMutationID(trimmedCommentID) {
		return false
	}

	updated := false
	filteredThreads := detail.InlineCommentThreads[:0]
	for _, thread := range detail.InlineCommentThreads {
		comments := thread.Comments[:0]
		removedFromThread := false
		for _, comment := range thread.Comments {
			if strings.TrimSpace(comment.ID) == trimmedCommentID {
				updated = true
				removedFromThread = true
				continue
			}
			comments = append(comments, comment)
		}
		thread.Comments = comments
		if len(thread.Comments) > 0 {
			filteredThreads = append(filteredThreads, thread)
			continue
		}
		if !removedFromThread {
			filteredThreads = append(filteredThreads, thread)
		}
	}
	detail.InlineCommentThreads = filteredThreads

	filteredInlineComments := detail.InlineComments[:0]
	for _, comment := range detail.InlineComments {
		if strings.TrimSpace(comment.ID) == trimmedCommentID {
			updated = true
			continue
		}
		filteredInlineComments = append(filteredInlineComments, comment)
	}
	detail.InlineComments = filteredInlineComments
	return updated
}

func deleteReviewCommentFromPullRequestDiff(data *reviewDiffData, commentID string) bool {
	trimmedCommentID := strings.TrimSpace(commentID)
	if data == nil || !hasUsablePullRequestMutationID(trimmedCommentID) {
		return false
	}

	updated := false
	for fileIndex := range data.Files {
		filteredThreads := data.Files[fileIndex].Threads[:0]
		for _, thread := range data.Files[fileIndex].Threads {
			comments := thread.Comments[:0]
			removedFromThread := false
			for _, comment := range thread.Comments {
				if strings.TrimSpace(comment.ID) == trimmedCommentID {
					updated = true
					removedFromThread = true
					continue
				}
				comments = append(comments, comment)
			}
			thread.Comments = comments
			if len(thread.Comments) > 0 {
				filteredThreads = append(filteredThreads, thread)
				continue
			}
			if !removedFromThread {
				filteredThreads = append(filteredThreads, thread)
			}
		}
		data.Files[fileIndex].Threads = filteredThreads
	}
	return updated
}
