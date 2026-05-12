package githubcli

import "strconv"

func (client *PullRequestMutationService) OpenPullRequestInBrowser(repository string, number int) error {
	trimmedRepository, err := normalizePullRequestIdentity(repository, number)
	if err != nil {
		return err
	}

	if _, err := client.execute(rawCommand("pr", "view", strconv.Itoa(number), "-R", trimmedRepository, "--web")); err != nil {
		return err
	}

	return nil
}
