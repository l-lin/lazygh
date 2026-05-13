package githubcli

type BuildAssembler struct {
	LoadBuildInfos func(repository string, number int) ([]PullRequestBuildInfo, error)
}

type pullRequestBuildInfoDTO PullRequestBuildInfo

func newBuildAssembler(service serviceBase) BuildAssembler {
	return BuildAssembler{LoadBuildInfos: service.listPullRequestBuildInfos}
}

func (BuildAssembler) ParseBuildInfos(stdout []byte) ([]PullRequestBuildInfo, error) {
	var dtos []pullRequestBuildInfoDTO
	if err := decodeEndpointJSONResponse(stdout, &dtos, ErrInvalidPullRequestBuildResponse); err != nil {
		return nil, err
	}
	return mapPullRequestBuildInfoDTOs(dtos), nil
}

func (assembler BuildAssembler) HydrateStatusCheckLinks(repository string, number int, checks []PullRequestStatusCheck) []PullRequestStatusCheck {
	if len(checks) == 0 {
		return nil
	}
	if assembler.LoadBuildInfos == nil {
		return mergePullRequestStatusCheckLinks(checks, nil)
	}

	buildInfos, err := assembler.LoadBuildInfos(repository, number)
	if err != nil {
		return mergePullRequestStatusCheckLinks(checks, nil)
	}
	return mergePullRequestStatusCheckLinks(checks, buildInfos)
}

func (assembler BuildAssembler) FindBuildInfo(repository string, number int, check PullRequestStatusCheck) (PullRequestBuildInfo, error) {
	if assembler.LoadBuildInfos == nil {
		return PullRequestBuildInfo{}, ErrPullRequestBuildInfoNotFound
	}

	buildInfos, err := assembler.LoadBuildInfos(repository, number)
	if err != nil {
		return PullRequestBuildInfo{}, err
	}
	actual, ok := pullRequestBuildInfoMatchingCheck(check, buildInfos)
	if !ok {
		return PullRequestBuildInfo{}, ErrPullRequestBuildInfoNotFound
	}
	return actual, nil
}

func mapPullRequestBuildInfoDTOs(dtos []pullRequestBuildInfoDTO) []PullRequestBuildInfo {
	if len(dtos) == 0 {
		return nil
	}

	buildInfos := make([]PullRequestBuildInfo, 0, len(dtos))
	for _, dto := range dtos {
		buildInfos = append(buildInfos, PullRequestBuildInfo(dto).normalized())
	}
	return buildInfos
}
