package githubcli

type PullRequestDetailAssembler struct {
	LoadBaseDetail                  func(repository string, number int) (PullRequestDetail, error)
	LoadMergeQueueMetadata          func(repository string, number int) (pullRequestMergeQueueMetadata, error)
	LoadOutOfDateWithBase           func(repository string, detail PullRequestDetail) (bool, error)
	HydrateBuildLinks               func(repository string, number int, checks []PullRequestStatusCheck) []PullRequestStatusCheck
	ListInlineComments              func(repository string, number int) ([]PullRequestInlineComment, error)
	ListReviewThreads               func(repository string, number int) ([]PullRequestReviewThread, error)
	ListReactionTargets             func(repository string, number int) (pullRequestReactionTargets, error)
	ListReviewCommentReactionGroups func(ids []string) (map[string][]ReactionGroup, error)
}

func newPullRequestDetailAssembler(service *PullRequestDetailService) PullRequestDetailAssembler {
	builds := newBuildAssembler(service.serviceBase)
	return PullRequestDetailAssembler{
		LoadBaseDetail:                  service.loadPullRequestBaseDetail,
		LoadMergeQueueMetadata:          service.loadPullRequestMergeQueueMetadata,
		LoadOutOfDateWithBase:           service.pullRequestOutOfDateWithBase,
		HydrateBuildLinks:               builds.HydrateStatusCheckLinks,
		ListInlineComments:              service.listPullRequestInlineComments,
		ListReviewThreads:               service.listPullRequestReviewThreads,
		ListReactionTargets:             service.listPullRequestReactionTargets,
		ListReviewCommentReactionGroups: service.listPullRequestReviewCommentReactionGroups,
	}
}

func (assembler PullRequestDetailAssembler) Assemble(repository string, number int) (PullRequestDetail, error) {
	if assembler.LoadBaseDetail == nil {
		return PullRequestDetail{}, ErrInvalidPullRequestDetailResponse
	}

	detail, err := assembler.LoadBaseDetail(repository, number)
	if err != nil {
		return PullRequestDetail{}, err
	}
	if assembler.LoadMergeQueueMetadata != nil {
		mergeQueueMetadata, actualErr := assembler.LoadMergeQueueMetadata(repository, number)
		if actualErr != nil {
			return PullRequestDetail{}, actualErr
		}
		detail = applyPullRequestMergeQueueMetadata(detail, mergeQueueMetadata)
	}
	if assembler.LoadOutOfDateWithBase != nil {
		outOfDateWithBase, actualErr := assembler.LoadOutOfDateWithBase(repository, detail)
		if actualErr == nil {
			detail.OutOfDateWithBase = outOfDateWithBase
		}
	}
	if assembler.HydrateBuildLinks != nil && len(detail.StatusCheckRollup) > 0 {
		detail.StatusCheckRollup = assembler.HydrateBuildLinks(repository, number, detail.StatusCheckRollup)
	}
	if assembler.ListInlineComments != nil {
		inlineComments, err := assembler.ListInlineComments(repository, number)
		if err != nil {
			return PullRequestDetail{}, err
		}
		if len(inlineComments) > 0 {
			detail.InlineComments = inlineComments
		}
	}
	if assembler.ListReviewThreads != nil {
		inlineThreads, err := assembler.ListReviewThreads(repository, number)
		if err != nil {
			return PullRequestDetail{}, err
		}
		if len(inlineThreads) > 0 {
			detail.InlineCommentThreads = inlineThreads
		}
	}
	if assembler.ListReactionTargets != nil {
		reactionTargets, err := assembler.ListReactionTargets(repository, number)
		if err != nil {
			return PullRequestDetail{}, err
		}
		if reactionTargets.PullRequestID != "" {
			detail.ID = reactionTargets.PullRequestID
		}
		if len(reactionTargets.ReactionGroups) > 0 {
			detail.ReactionGroups = reactionTargets.ReactionGroups
		}
		if len(reactionTargets.Comments) > 0 {
			detail.Comments = reactionTargets.Comments
		}
	}
	if assembler.ListReviewCommentReactionGroups != nil {
		reactionTargetIDs := append(pullRequestInlineCommentReactionTargetIDs(detail.InlineComments), pullRequestReviewReactionTargetIDs(detail.Reviews)...)
		reactionGroupsByID, err := assembler.ListReviewCommentReactionGroups(reactionTargetIDs)
		if err != nil {
			return PullRequestDetail{}, err
		}
		if len(reactionGroupsByID) > 0 {
			detail.InlineComments = mergePullRequestInlineCommentReactionGroups(detail.InlineComments, reactionGroupsByID)
			detail.Reviews = mergePullRequestReviewReactionGroups(detail.Reviews, reactionGroupsByID)
		}
	}

	return detail.normalized(), nil
}
