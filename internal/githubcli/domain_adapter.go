package githubcli

import githubdomain "github.com/l-lin/lazygh/internal/github"

func ToDomainPullRequests(pullRequests []PullRequest) []githubdomain.PullRequestSummary {
	if len(pullRequests) == 0 {
		return nil
	}

	converted := make([]githubdomain.PullRequestSummary, 0, len(pullRequests))
	for _, pullRequest := range pullRequests {
		converted = append(converted, ToDomainPullRequestSummary(pullRequest))
	}
	return converted
}

func PullRequestsFromDomain(pullRequests []githubdomain.PullRequestSummary) []PullRequest {
	if len(pullRequests) == 0 {
		return nil
	}

	converted := make([]PullRequest, 0, len(pullRequests))
	for _, pullRequest := range pullRequests {
		converted = append(converted, PullRequestSummaryFromDomain(pullRequest))
	}
	return converted
}

func ToDomainPullRequestSummary(pullRequest PullRequest) githubdomain.PullRequestSummary {
	return githubdomain.PullRequestSummary{
		ID:                     pullRequest.ID,
		Title:                  pullRequest.Title,
		Number:                 pullRequest.Number,
		Repository:             ToDomainRepository(pullRequest.Repository),
		URL:                    pullRequest.URL,
		Body:                   pullRequest.Body,
		State:                  pullRequest.State,
		IsDraft:                pullRequest.IsDraft,
		UpdatedAt:              pullRequest.UpdatedAt,
		ReviewDecision:         pullRequest.ReviewDecision,
		ReviewRequests:         toDomainReviewRequests(pullRequest.ReviewRequests),
		MergeStateStatus:       pullRequest.MergeStateStatus,
		Mergeable:              pullRequest.Mergeable,
		AutoMergeRequest:       toDomainPullRequestAutoMergeRequest(pullRequest.AutoMergeRequest),
		StatusCheckRollupState: pullRequest.StatusCheckRollupState,
	}
}

func PullRequestSummaryFromDomain(pullRequest githubdomain.PullRequestSummary) PullRequest {
	return PullRequest{
		ID:                     pullRequest.ID,
		Title:                  pullRequest.Title,
		Number:                 pullRequest.Number,
		Repository:             RepositoryFromDomain(pullRequest.Repository),
		URL:                    pullRequest.URL,
		Body:                   pullRequest.Body,
		State:                  pullRequest.State,
		IsDraft:                pullRequest.IsDraft,
		UpdatedAt:              pullRequest.UpdatedAt,
		ReviewDecision:         pullRequest.ReviewDecision,
		ReviewRequests:         reviewRequestsFromDomain(pullRequest.ReviewRequests),
		MergeStateStatus:       pullRequest.MergeStateStatus,
		Mergeable:              pullRequest.Mergeable,
		AutoMergeRequest:       pullRequestAutoMergeRequestFromDomain(pullRequest.AutoMergeRequest),
		StatusCheckRollupState: pullRequest.StatusCheckRollupState,
	}
}

func ToDomainRepository(repository Repository) githubdomain.RepositoryRef {
	return githubdomain.RepositoryRef{Name: repository.Name, NameWithOwner: repository.NameWithOwner}
}

func RepositoryFromDomain(repository githubdomain.RepositoryRef) Repository {
	return Repository{Name: repository.Name, NameWithOwner: repository.NameWithOwner}
}

func toDomainPullRequestAutoMergeRequest(request *PullRequestAutoMergeRequest) *githubdomain.PullRequestAutoMergeRequest {
	if request == nil {
		return nil
	}
	actual := githubdomain.PullRequestAutoMergeRequest{EnabledAt: request.EnabledAt}
	return &actual
}

func pullRequestAutoMergeRequestFromDomain(request *githubdomain.PullRequestAutoMergeRequest) *PullRequestAutoMergeRequest {
	if request == nil {
		return nil
	}
	actual := PullRequestAutoMergeRequest{EnabledAt: request.EnabledAt}
	return &actual
}

func toDomainReviewRequests(reviewRequests []PullRequestReviewRequest) []githubdomain.PullRequestReviewRequest {
	if len(reviewRequests) == 0 {
		return nil
	}

	converted := make([]githubdomain.PullRequestReviewRequest, 0, len(reviewRequests))
	for _, reviewRequest := range reviewRequests {
		converted = append(converted, githubdomain.PullRequestReviewRequest{RequestedReviewer: toDomainRequestedReviewer(reviewRequest.RequestedReviewer)})
	}
	return converted
}

func reviewRequestsFromDomain(reviewRequests []githubdomain.PullRequestReviewRequest) []PullRequestReviewRequest {
	if len(reviewRequests) == 0 {
		return nil
	}

	converted := make([]PullRequestReviewRequest, 0, len(reviewRequests))
	for _, reviewRequest := range reviewRequests {
		converted = append(converted, PullRequestReviewRequest{RequestedReviewer: requestedReviewerFromDomain(reviewRequest.RequestedReviewer)})
	}
	return converted
}

func toDomainRequestedReviewer(reviewer PullRequestRequestedReviewer) githubdomain.PullRequestRequestedReviewer {
	actual := githubdomain.PullRequestRequestedReviewer{TypeName: reviewer.TypeName, Login: reviewer.Login, Name: reviewer.Name, Slug: reviewer.Slug}
	if reviewer.Organization != nil {
		actual.Organization = &githubdomain.PullRequestReviewRequestOrganization{Login: reviewer.Organization.Login}
	}
	return actual
}

func requestedReviewerFromDomain(reviewer githubdomain.PullRequestRequestedReviewer) PullRequestRequestedReviewer {
	actual := PullRequestRequestedReviewer{TypeName: reviewer.TypeName, Login: reviewer.Login, Name: reviewer.Name, Slug: reviewer.Slug}
	if reviewer.Organization != nil {
		actual.Organization = &PullRequestReviewRequestOrganization{Login: reviewer.Organization.Login}
	}
	return actual
}

func ToDomainPullRequestDetail(detail PullRequestDetail) githubdomain.PullRequestDetail {
	actual := githubdomain.PullRequestDetail{
		ID:                   detail.ID,
		Title:                detail.Title,
		Number:               detail.Number,
		URL:                  detail.URL,
		Body:                 detail.Body,
		BodyHTML:             detail.BodyHTML,
		State:                detail.State,
		IsDraft:              detail.IsDraft,
		CreatedAt:            detail.CreatedAt,
		UpdatedAt:            detail.UpdatedAt,
		Labels:               toDomainPullRequestLabels(detail.Labels),
		Assignees:            toDomainPullRequestAuthors(detail.Assignees),
		ReviewDecision:       detail.ReviewDecision,
		ReviewRequests:       toDomainReviewRequests(detail.ReviewRequests),
		BaseRefName:          detail.BaseRefName,
		HeadRefName:          detail.HeadRefName,
		MergeStateStatus:     detail.MergeStateStatus,
		Mergeable:            detail.Mergeable,
		AutoMergeRequest:     toDomainPullRequestAutoMergeRequest(detail.AutoMergeRequest),
		OutOfDateWithBase:    detail.OutOfDateWithBase,
		ReactionGroups:       toDomainReactionGroups(detail.ReactionGroups),
		Comments:             toDomainPullRequestComments(detail.Comments),
		Commits:              toDomainPullRequestCommits(detail.Commits),
		Reviews:              toDomainPullRequestReviews(detail.Reviews),
		InlineComments:       toDomainPullRequestInlineComments(detail.InlineComments),
		InlineCommentThreads: toDomainReviewThreads(detail.InlineCommentThreads),
		Additions:            detail.Additions,
		Deletions:            detail.Deletions,
		ChangedFiles:         detail.ChangedFiles,
		StatusCheckRollup:    toDomainBuildInfos(detail.StatusCheckRollup),
	}
	if detail.Author != nil {
		author := toDomainPullRequestAuthor(*detail.Author)
		actual.Author = &author
	}
	return actual
}

func PullRequestDetailFromDomain(detail githubdomain.PullRequestDetail) PullRequestDetail {
	actual := PullRequestDetail{
		ID:                   detail.ID,
		Title:                detail.Title,
		Number:               detail.Number,
		URL:                  detail.URL,
		Body:                 detail.Body,
		BodyHTML:             detail.BodyHTML,
		State:                detail.State,
		IsDraft:              detail.IsDraft,
		CreatedAt:            detail.CreatedAt,
		UpdatedAt:            detail.UpdatedAt,
		Labels:               pullRequestLabelsFromDomain(detail.Labels),
		Assignees:            pullRequestAuthorsFromDomain(detail.Assignees),
		ReviewDecision:       detail.ReviewDecision,
		ReviewRequests:       reviewRequestsFromDomain(detail.ReviewRequests),
		BaseRefName:          detail.BaseRefName,
		HeadRefName:          detail.HeadRefName,
		MergeStateStatus:     detail.MergeStateStatus,
		Mergeable:            detail.Mergeable,
		AutoMergeRequest:     pullRequestAutoMergeRequestFromDomain(detail.AutoMergeRequest),
		OutOfDateWithBase:    detail.OutOfDateWithBase,
		ReactionGroups:       reactionGroupsFromDomain(detail.ReactionGroups),
		Comments:             pullRequestCommentsFromDomain(detail.Comments),
		Commits:              pullRequestCommitsFromDomain(detail.Commits),
		Reviews:              pullRequestReviewsFromDomain(detail.Reviews),
		InlineComments:       pullRequestInlineCommentsFromDomain(detail.InlineComments),
		InlineCommentThreads: reviewThreadsFromDomain(detail.InlineCommentThreads),
		Additions:            detail.Additions,
		Deletions:            detail.Deletions,
		ChangedFiles:         detail.ChangedFiles,
		StatusCheckRollup:    buildInfosFromDomain(detail.StatusCheckRollup),
	}
	if detail.Author != nil {
		author := pullRequestAuthorFromDomain(*detail.Author)
		actual.Author = &author
	}
	return actual
}

func ToDomainPullRequestDiff(diff PullRequestDiff) githubdomain.PullRequestDiff {
	return githubdomain.PullRequestDiff{
		UnifiedDiff:             diff.UnifiedDiff,
		Files:                   toDomainPullRequestDiffFiles(diff.Files),
		Threads:                 toDomainReviewThreads(diff.Threads),
		FileTeamOwnersAttempted: diff.FileTeamOwnersAttempted,
	}
}

func PullRequestDiffFromDomain(diff githubdomain.PullRequestDiff) PullRequestDiff {
	return PullRequestDiff{
		UnifiedDiff:             diff.UnifiedDiff,
		Files:                   pullRequestDiffFilesFromDomain(diff.Files),
		Threads:                 reviewThreadsFromDomain(diff.Threads),
		FileTeamOwnersAttempted: diff.FileTeamOwnersAttempted,
	}
}

func ToDomainNotifications(notifications []Notification) []githubdomain.Notification {
	if len(notifications) == 0 {
		return nil
	}

	converted := make([]githubdomain.Notification, 0, len(notifications))
	for _, notification := range notifications {
		converted = append(converted, ToDomainNotification(notification))
	}
	return converted
}

func NotificationsFromDomain(notifications []githubdomain.Notification) []Notification {
	if len(notifications) == 0 {
		return nil
	}

	converted := make([]Notification, 0, len(notifications))
	for _, notification := range notifications {
		converted = append(converted, NotificationFromDomain(notification))
	}
	return converted
}

func ToDomainNotification(notification Notification) githubdomain.Notification {
	return githubdomain.Notification{
		ID:              notification.ID,
		Done:            notification.Done,
		Unread:          notification.Unread,
		Reason:          notification.Reason,
		UpdatedAt:       notification.UpdatedAt,
		LastReadAt:      notification.LastReadAt,
		URL:             notification.URL,
		SubscriptionURL: notification.SubscriptionURL,
		Repository:      ToDomainRepository(notification.Repository),
		Subject:         githubdomain.NotificationSubject{Title: notification.Subject.Title, Type: notification.Subject.Type, URL: notification.Subject.URL, LatestCommentURL: notification.Subject.LatestCommentURL},
	}
}

func NotificationFromDomain(notification githubdomain.Notification) Notification {
	return Notification{
		ID:              notification.ID,
		Done:            notification.Done,
		Unread:          notification.Unread,
		Reason:          notification.Reason,
		UpdatedAt:       notification.UpdatedAt,
		LastReadAt:      notification.LastReadAt,
		URL:             notification.URL,
		SubscriptionURL: notification.SubscriptionURL,
		Repository:      RepositoryFromDomain(notification.Repository),
		Subject:         NotificationSubject{Title: notification.Subject.Title, Type: notification.Subject.Type, URL: notification.Subject.URL, LatestCommentURL: notification.Subject.LatestCommentURL},
	}
}

func ToDomainIssueDetail(detail IssueDetail) githubdomain.IssueDetail {
	actual := githubdomain.IssueDetail{
		Title:     detail.Title,
		Number:    detail.Number,
		URL:       detail.URL,
		Body:      detail.Body,
		BodyHTML:  detail.BodyHTML,
		State:     detail.State,
		CreatedAt: detail.CreatedAt,
		UpdatedAt: detail.UpdatedAt,
		Labels:    toDomainPullRequestLabels(detail.Labels),
		Assignees: toDomainPullRequestAuthors(detail.Assignees),
		Comments:  detail.Comments,
	}
	if detail.Author != nil {
		author := toDomainPullRequestAuthor(*detail.Author)
		actual.Author = &author
	}
	return actual
}

func IssueDetailFromDomain(detail githubdomain.IssueDetail) IssueDetail {
	actual := IssueDetail{
		Title:     detail.Title,
		Number:    detail.Number,
		URL:       detail.URL,
		Body:      detail.Body,
		BodyHTML:  detail.BodyHTML,
		State:     detail.State,
		CreatedAt: detail.CreatedAt,
		UpdatedAt: detail.UpdatedAt,
		Labels:    pullRequestLabelsFromDomain(detail.Labels),
		Assignees: pullRequestAuthorsFromDomain(detail.Assignees),
		Comments:  detail.Comments,
	}
	if detail.Author != nil {
		author := pullRequestAuthorFromDomain(*detail.Author)
		actual.Author = &author
	}
	return actual
}

func ToDomainReleaseDetail(detail ReleaseDetail) githubdomain.ReleaseDetail {
	actual := githubdomain.ReleaseDetail{
		Name:        detail.Name,
		TagName:     detail.TagName,
		URL:         detail.URL,
		Body:        detail.Body,
		BodyHTML:    detail.BodyHTML,
		Draft:       detail.Draft,
		PreRelease:  detail.PreRelease,
		CreatedAt:   detail.CreatedAt,
		UpdatedAt:   detail.UpdatedAt,
		PublishedAt: detail.PublishedAt,
	}
	if detail.Author != nil {
		author := toDomainPullRequestAuthor(*detail.Author)
		actual.Author = &author
	}
	return actual
}

func ReleaseDetailFromDomain(detail githubdomain.ReleaseDetail) ReleaseDetail {
	actual := ReleaseDetail{
		Name:        detail.Name,
		TagName:     detail.TagName,
		URL:         detail.URL,
		Body:        detail.Body,
		BodyHTML:    detail.BodyHTML,
		Draft:       detail.Draft,
		PreRelease:  detail.PreRelease,
		CreatedAt:   detail.CreatedAt,
		UpdatedAt:   detail.UpdatedAt,
		PublishedAt: detail.PublishedAt,
	}
	if detail.Author != nil {
		author := pullRequestAuthorFromDomain(*detail.Author)
		actual.Author = &author
	}
	return actual
}

func ToDomainConnectedUser(user ConnectedUser) githubdomain.ConnectedUser {
	return githubdomain.ConnectedUser{
		Login:       user.Login,
		Name:        user.Name,
		Bio:         user.Bio,
		Company:     user.Company,
		Location:    user.Location,
		PublicRepos: user.PublicRepos,
		Followers:   user.Followers,
		URL:         user.URL,
	}
}

func ConnectedUserFromDomain(user githubdomain.ConnectedUser) ConnectedUser {
	return ConnectedUser{
		Login:       user.Login,
		Name:        user.Name,
		Bio:         user.Bio,
		Company:     user.Company,
		Location:    user.Location,
		PublicRepos: user.PublicRepos,
		Followers:   user.Followers,
		URL:         user.URL,
	}
}

func ToDomainNotificationBulkReadResult(result NotificationBulkReadResult) githubdomain.NotificationBulkReadResult {
	return githubdomain.NotificationBulkReadResult{Accepted: result.Accepted}
}

func NotificationBulkReadResultFromDomain(result githubdomain.NotificationBulkReadResult) NotificationBulkReadResult {
	return NotificationBulkReadResult{Accepted: result.Accepted}
}

func ToDomainPullRequestAuthor(author PullRequestAuthor) githubdomain.PullRequestAuthor {
	return toDomainPullRequestAuthor(author)
}

func PullRequestAuthorFromDomain(author githubdomain.PullRequestAuthor) PullRequestAuthor {
	return pullRequestAuthorFromDomain(author)
}

func ToDomainPullRequestAuthors(authors []PullRequestAuthor) []githubdomain.PullRequestAuthor {
	return toDomainPullRequestAuthors(authors)
}

func PullRequestAuthorsFromDomain(authors []githubdomain.PullRequestAuthor) []PullRequestAuthor {
	return pullRequestAuthorsFromDomain(authors)
}

func ToDomainPullRequestReviewEvent(event PullRequestReviewEvent) githubdomain.PullRequestReviewEvent {
	return githubdomain.PullRequestReviewEvent(event)
}

func PullRequestReviewEventFromDomain(event githubdomain.PullRequestReviewEvent) PullRequestReviewEvent {
	return PullRequestReviewEvent(event)
}

func ToDomainPullRequestReviewThreadTarget(target PullRequestReviewThreadTarget) githubdomain.PullRequestReviewThreadTarget {
	return githubdomain.ReviewThreadTarget{
		Path:        target.Path,
		Line:        target.Line,
		Side:        target.Side,
		StartLine:   target.StartLine,
		StartSide:   target.StartSide,
		SubjectType: target.SubjectType,
	}
}

func PullRequestReviewThreadTargetFromDomain(target githubdomain.PullRequestReviewThreadTarget) PullRequestReviewThreadTarget {
	return PullRequestReviewThreadTarget{
		Path:        target.Path,
		Line:        target.Line,
		Side:        target.Side,
		StartLine:   target.StartLine,
		StartSide:   target.StartSide,
		SubjectType: target.SubjectType,
	}
}

func ToDomainReactionContent(content ReactionContent) githubdomain.ReactionContent {
	return githubdomain.ReactionContent(content)
}

func ReactionContentFromDomain(content githubdomain.ReactionContent) ReactionContent {
	return ReactionContent(content)
}

func ToDomainPullRequestStatusCheck(check PullRequestStatusCheck) githubdomain.PullRequestStatusCheck {
	return githubdomain.BuildInfo{
		TypeName:     check.TypeName,
		Name:         check.Name,
		Status:       check.Status,
		Conclusion:   check.Conclusion,
		WorkflowName: check.WorkflowName,
		Link:         check.Link,
	}
}

func PullRequestStatusCheckFromDomain(check githubdomain.PullRequestStatusCheck) PullRequestStatusCheck {
	return PullRequestStatusCheck{
		TypeName:     check.TypeName,
		Name:         check.Name,
		Status:       check.Status,
		Conclusion:   check.Conclusion,
		WorkflowName: check.WorkflowName,
		Link:         check.Link,
	}
}

func ToDomainPullRequestBuildRunJob(job PullRequestBuildRunJob) githubdomain.PullRequestBuildRunJob {
	return githubdomain.BuildRunJob{
		DatabaseID: job.DatabaseID,
		Name:       job.Name,
		Status:     job.Status,
		Conclusion: job.Conclusion,
		URL:        job.URL,
	}
}

func PullRequestBuildRunJobFromDomain(job githubdomain.PullRequestBuildRunJob) PullRequestBuildRunJob {
	return PullRequestBuildRunJob{
		DatabaseID: job.DatabaseID,
		Name:       job.Name,
		Status:     job.Status,
		Conclusion: job.Conclusion,
		URL:        job.URL,
	}
}

func ToDomainPullRequestBuildRunJobs(jobs []PullRequestBuildRunJob) []githubdomain.PullRequestBuildRunJob {
	if len(jobs) == 0 {
		return nil
	}

	converted := make([]githubdomain.PullRequestBuildRunJob, 0, len(jobs))
	for _, job := range jobs {
		converted = append(converted, ToDomainPullRequestBuildRunJob(job))
	}
	return converted
}

func PullRequestBuildRunJobsFromDomain(jobs []githubdomain.PullRequestBuildRunJob) []PullRequestBuildRunJob {
	if len(jobs) == 0 {
		return nil
	}

	converted := make([]PullRequestBuildRunJob, 0, len(jobs))
	for _, job := range jobs {
		converted = append(converted, PullRequestBuildRunJobFromDomain(job))
	}
	return converted
}

func toDomainPullRequestAuthors(authors []PullRequestAuthor) []githubdomain.PullRequestAuthor {
	if len(authors) == 0 {
		return nil
	}
	converted := make([]githubdomain.PullRequestAuthor, 0, len(authors))
	for _, author := range authors {
		converted = append(converted, toDomainPullRequestAuthor(author))
	}
	return converted
}

func pullRequestAuthorsFromDomain(authors []githubdomain.PullRequestAuthor) []PullRequestAuthor {
	if len(authors) == 0 {
		return nil
	}
	converted := make([]PullRequestAuthor, 0, len(authors))
	for _, author := range authors {
		converted = append(converted, pullRequestAuthorFromDomain(author))
	}
	return converted
}

func toDomainPullRequestAuthor(author PullRequestAuthor) githubdomain.PullRequestAuthor {
	return githubdomain.PullRequestAuthor{Login: author.Login, Name: author.Name, IsBot: author.IsBot}
}

func pullRequestAuthorFromDomain(author githubdomain.PullRequestAuthor) PullRequestAuthor {
	return PullRequestAuthor{Login: author.Login, Name: author.Name, IsBot: author.IsBot}
}

func toDomainPullRequestLabels(labels []PullRequestLabel) []githubdomain.PullRequestLabel {
	if len(labels) == 0 {
		return nil
	}
	converted := make([]githubdomain.PullRequestLabel, 0, len(labels))
	for _, label := range labels {
		converted = append(converted, githubdomain.PullRequestLabel{Name: label.Name})
	}
	return converted
}

func pullRequestLabelsFromDomain(labels []githubdomain.PullRequestLabel) []PullRequestLabel {
	if len(labels) == 0 {
		return nil
	}
	converted := make([]PullRequestLabel, 0, len(labels))
	for _, label := range labels {
		converted = append(converted, PullRequestLabel{Name: label.Name})
	}
	return converted
}

func toDomainPullRequestComments(comments []PullRequestComment) []githubdomain.PullRequestComment {
	if len(comments) == 0 {
		return nil
	}
	converted := make([]githubdomain.PullRequestComment, 0, len(comments))
	for _, comment := range comments {
		actual := githubdomain.PullRequestComment{ID: comment.ID, Body: comment.Body, BodyHTML: comment.BodyHTML, CreatedAt: comment.CreatedAt, URL: comment.URL, DiffHunk: comment.DiffHunk, State: comment.State, ViewerDidAuthor: comment.ViewerDidAuthor, ReactionGroups: toDomainReactionGroups(comment.ReactionGroups)}
		if comment.Author != nil {
			author := githubdomain.PullRequestCommentAuthor{Login: comment.Author.Login}
			actual.Author = &author
		}
		converted = append(converted, actual)
	}
	return converted
}

func pullRequestCommentsFromDomain(comments []githubdomain.PullRequestComment) []PullRequestComment {
	if len(comments) == 0 {
		return nil
	}
	converted := make([]PullRequestComment, 0, len(comments))
	for _, comment := range comments {
		actual := PullRequestComment{ID: comment.ID, Body: comment.Body, BodyHTML: comment.BodyHTML, CreatedAt: comment.CreatedAt, URL: comment.URL, DiffHunk: comment.DiffHunk, State: comment.State, ViewerDidAuthor: comment.ViewerDidAuthor, ReactionGroups: reactionGroupsFromDomain(comment.ReactionGroups)}
		if comment.Author != nil {
			author := PullRequestCommentAuthor{Login: comment.Author.Login}
			actual.Author = &author
		}
		converted = append(converted, actual)
	}
	return converted
}

func toDomainPullRequestCommits(commits []PullRequestCommit) []githubdomain.PullRequestCommit {
	if len(commits) == 0 {
		return nil
	}
	converted := make([]githubdomain.PullRequestCommit, 0, len(commits))
	for _, commit := range commits {
		converted = append(converted, githubdomain.PullRequestCommit{OID: commit.OID, MessageHeadline: commit.MessageHeadline, MessageBody: commit.MessageBody, MessageBodyHTML: commit.MessageBodyHTML, AuthoredDate: commit.AuthoredDate, CommittedDate: commit.CommittedDate, Authors: toDomainCommitAuthors(commit.Authors)})
	}
	return converted
}

func pullRequestCommitsFromDomain(commits []githubdomain.PullRequestCommit) []PullRequestCommit {
	if len(commits) == 0 {
		return nil
	}
	converted := make([]PullRequestCommit, 0, len(commits))
	for _, commit := range commits {
		converted = append(converted, PullRequestCommit{OID: commit.OID, MessageHeadline: commit.MessageHeadline, MessageBody: commit.MessageBody, MessageBodyHTML: commit.MessageBodyHTML, AuthoredDate: commit.AuthoredDate, CommittedDate: commit.CommittedDate, Authors: commitAuthorsFromDomain(commit.Authors)})
	}
	return converted
}

func toDomainCommitAuthors(authors []PullRequestCommitAuthor) []githubdomain.PullRequestCommitAuthor {
	if len(authors) == 0 {
		return nil
	}
	converted := make([]githubdomain.PullRequestCommitAuthor, 0, len(authors))
	for _, author := range authors {
		converted = append(converted, githubdomain.PullRequestCommitAuthor{Login: author.Login, Name: author.Name, Email: author.Email})
	}
	return converted
}

func commitAuthorsFromDomain(authors []githubdomain.PullRequestCommitAuthor) []PullRequestCommitAuthor {
	if len(authors) == 0 {
		return nil
	}
	converted := make([]PullRequestCommitAuthor, 0, len(authors))
	for _, author := range authors {
		converted = append(converted, PullRequestCommitAuthor{Login: author.Login, Name: author.Name, Email: author.Email})
	}
	return converted
}

func toDomainPullRequestInlineComments(comments []PullRequestInlineComment) []githubdomain.PullRequestInlineComment {
	if len(comments) == 0 {
		return nil
	}
	converted := make([]githubdomain.PullRequestInlineComment, 0, len(comments))
	for _, comment := range comments {
		actual := githubdomain.PullRequestInlineComment{ID: comment.ID, Body: comment.Body, BodyHTML: comment.BodyHTML, CreatedAt: comment.CreatedAt, URL: comment.URL, Path: comment.Path, DiffHunk: comment.DiffHunk, Line: comment.Line, OriginalLine: comment.OriginalLine, StartLine: comment.StartLine, OriginalStartLine: comment.OriginalStartLine, Side: comment.Side, StartSide: comment.StartSide, SubjectType: comment.SubjectType, ReactionGroups: toDomainReactionGroups(comment.ReactionGroups)}
		if comment.Author != nil {
			author := githubdomain.PullRequestCommentAuthor{Login: comment.Author.Login}
			actual.Author = &author
		}
		converted = append(converted, actual)
	}
	return converted
}

func pullRequestInlineCommentsFromDomain(comments []githubdomain.PullRequestInlineComment) []PullRequestInlineComment {
	if len(comments) == 0 {
		return nil
	}
	converted := make([]PullRequestInlineComment, 0, len(comments))
	for _, comment := range comments {
		actual := PullRequestInlineComment{ID: comment.ID, Body: comment.Body, BodyHTML: comment.BodyHTML, CreatedAt: comment.CreatedAt, URL: comment.URL, Path: comment.Path, DiffHunk: comment.DiffHunk, Line: comment.Line, OriginalLine: comment.OriginalLine, StartLine: comment.StartLine, OriginalStartLine: comment.OriginalStartLine, Side: comment.Side, StartSide: comment.StartSide, SubjectType: comment.SubjectType, ReactionGroups: reactionGroupsFromDomain(comment.ReactionGroups)}
		if comment.Author != nil {
			author := PullRequestCommentAuthor{Login: comment.Author.Login}
			actual.Author = &author
		}
		converted = append(converted, actual)
	}
	return converted
}

func toDomainPullRequestReviews(reviews []PullRequestReview) []githubdomain.PullRequestReview {
	if len(reviews) == 0 {
		return nil
	}
	converted := make([]githubdomain.PullRequestReview, 0, len(reviews))
	for _, review := range reviews {
		actual := githubdomain.PullRequestReview{State: review.State, SubmittedAt: review.SubmittedAt}
		if review.Author != nil {
			author := githubdomain.PullRequestCommentAuthor{Login: review.Author.Login}
			actual.Author = &author
		}
		converted = append(converted, actual)
	}
	return converted
}

func pullRequestReviewsFromDomain(reviews []githubdomain.PullRequestReview) []PullRequestReview {
	if len(reviews) == 0 {
		return nil
	}
	converted := make([]PullRequestReview, 0, len(reviews))
	for _, review := range reviews {
		actual := PullRequestReview{State: review.State, SubmittedAt: review.SubmittedAt}
		if review.Author != nil {
			author := PullRequestCommentAuthor{Login: review.Author.Login}
			actual.Author = &author
		}
		converted = append(converted, actual)
	}
	return converted
}

func toDomainBuildInfos(checks []PullRequestStatusCheck) []githubdomain.BuildInfo {
	if len(checks) == 0 {
		return nil
	}
	converted := make([]githubdomain.BuildInfo, 0, len(checks))
	for _, check := range checks {
		converted = append(converted, githubdomain.BuildInfo{TypeName: check.TypeName, Name: check.Name, Status: check.Status, Conclusion: check.Conclusion, WorkflowName: check.WorkflowName, Link: check.Link})
	}
	return converted
}

func buildInfosFromDomain(checks []githubdomain.BuildInfo) []PullRequestStatusCheck {
	if len(checks) == 0 {
		return nil
	}
	converted := make([]PullRequestStatusCheck, 0, len(checks))
	for _, check := range checks {
		converted = append(converted, PullRequestStatusCheck{TypeName: check.TypeName, Name: check.Name, Status: check.Status, Conclusion: check.Conclusion, WorkflowName: check.WorkflowName, Link: check.Link})
	}
	return converted
}

func toDomainReactionGroups(groups []ReactionGroup) []githubdomain.ReactionGroup {
	if len(groups) == 0 {
		return nil
	}
	converted := make([]githubdomain.ReactionGroup, 0, len(groups))
	for _, group := range groups {
		converted = append(converted, githubdomain.ReactionGroup{Content: githubdomain.ReactionContent(group.Content), TotalCount: group.TotalCount, ViewerHasReacted: group.ViewerHasReacted})
	}
	return converted
}

func reactionGroupsFromDomain(groups []githubdomain.ReactionGroup) []ReactionGroup {
	if len(groups) == 0 {
		return nil
	}
	converted := make([]ReactionGroup, 0, len(groups))
	for _, group := range groups {
		converted = append(converted, ReactionGroup{Content: ReactionContent(group.Content), TotalCount: group.TotalCount, ViewerHasReacted: group.ViewerHasReacted})
	}
	return converted
}

func toDomainReviewThreads(threads []PullRequestReviewThread) []githubdomain.ReviewThread {
	if len(threads) == 0 {
		return nil
	}
	converted := make([]githubdomain.ReviewThread, 0, len(threads))
	for _, thread := range threads {
		converted = append(converted, githubdomain.ReviewThread{ID: thread.ID, IsResolved: thread.IsResolved, IsOutdated: thread.IsOutdated, ViewerCanResolve: thread.ViewerCanResolve, ViewerCanUnresolve: thread.ViewerCanUnresolve, Path: thread.Path, Line: thread.Line, OriginalLine: thread.OriginalLine, StartLine: thread.StartLine, OriginalStartLine: thread.OriginalStartLine, DiffSide: thread.DiffSide, StartDiffSide: thread.StartDiffSide, Comments: toDomainPullRequestComments(thread.Comments)})
	}
	return converted
}

func reviewThreadsFromDomain(threads []githubdomain.ReviewThread) []PullRequestReviewThread {
	if len(threads) == 0 {
		return nil
	}
	converted := make([]PullRequestReviewThread, 0, len(threads))
	for _, thread := range threads {
		converted = append(converted, PullRequestReviewThread{ID: thread.ID, IsResolved: thread.IsResolved, IsOutdated: thread.IsOutdated, ViewerCanResolve: thread.ViewerCanResolve, ViewerCanUnresolve: thread.ViewerCanUnresolve, Path: thread.Path, Line: thread.Line, OriginalLine: thread.OriginalLine, StartLine: thread.StartLine, OriginalStartLine: thread.OriginalStartLine, DiffSide: thread.DiffSide, StartDiffSide: thread.StartDiffSide, Comments: pullRequestCommentsFromDomain(thread.Comments)})
	}
	return converted
}

func toDomainPullRequestDiffFiles(files []PullRequestDiffFile) []githubdomain.PullRequestDiffFile {
	if len(files) == 0 {
		return nil
	}
	converted := make([]githubdomain.PullRequestDiffFile, 0, len(files))
	for _, file := range files {
		converted = append(converted, githubdomain.PullRequestDiffFile{Path: file.Path, PreviousPath: file.PreviousPath, ChangeType: file.ChangeType, Additions: file.Additions, Deletions: file.Deletions, Patch: file.Patch, TeamOwners: append([]string(nil), file.TeamOwners...)})
	}
	return converted
}

func pullRequestDiffFilesFromDomain(files []githubdomain.PullRequestDiffFile) []PullRequestDiffFile {
	if len(files) == 0 {
		return nil
	}
	converted := make([]PullRequestDiffFile, 0, len(files))
	for _, file := range files {
		converted = append(converted, PullRequestDiffFile{Path: file.Path, PreviousPath: file.PreviousPath, ChangeType: file.ChangeType, Additions: file.Additions, Deletions: file.Deletions, Patch: file.Patch, TeamOwners: append([]string(nil), file.TeamOwners...)})
	}
	return converted
}
