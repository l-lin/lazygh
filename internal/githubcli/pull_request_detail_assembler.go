package githubcli

type PullRequestDetailAssembler struct {
	LoadBaseDetail                  func(repository string, number int) (PullRequestDetail, error)
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
		inlineCommentReactionGroups, err := assembler.ListReviewCommentReactionGroups(pullRequestInlineCommentReactionTargetIDs(detail.InlineComments))
		if err != nil {
			return PullRequestDetail{}, err
		}
		if len(inlineCommentReactionGroups) > 0 {
			detail.InlineComments = mergePullRequestInlineCommentReactionGroups(detail.InlineComments, inlineCommentReactionGroups)
		}
	}

	return detail.normalized(), nil
}
