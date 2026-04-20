package githubcli

func (client *Client) runGH(commandName string, args ...string) (CommandResult, error) {
	result, err := client.runner.Run(ghBinaryName, args...)
	if err != nil {
		return CommandResult{}, classifyCommandError(commandName, err, result.Stderr)
	}

	return result, nil
}

func (client *Client) runGHWithInput(commandName string, input []byte, args ...string) (CommandResult, error) {
	result, err := client.runner.RunWithInput(ghBinaryName, input, args...)
	if err != nil {
		return CommandResult{}, classifyCommandError(commandName, err, result.Stderr)
	}

	return result, nil
}
