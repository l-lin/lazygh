package githubcli

func rawCommand(args ...string) Command {
	return Command{Args: append([]string(nil), args...), DisplayArgs: append([]string(nil), args...)}
}

func rawCommandWithInput(input []byte, args ...string) Command {
	command := rawCommand(args...)
	if input != nil {
		command.Stdin = append([]byte(nil), input...)
	}
	return command
}

func formatCommand(args ...string) string {
	return commandFormatter{}.Format(rawCommand(args...))
}

func formatCommandArguments(args []string) string {
	return commandFormatter{}.Format(Command{Args: append([]string(nil), args...), DisplayArgs: append([]string(nil), args...)})
}

func (client *Client) execute(command Command) (CommandResult, error) {
	return client.transport.executor.Execute(command)
}

func (client *Client) queryGraphQL(request GraphQLRequest) (CommandResult, error) {
	return client.transport.graphql.Query(request)
}

func (client *Client) doREST(request RESTRequest) (CommandResult, error) {
	return client.transport.rest.Do(request)
}

func (client *Client) decodePaginatedOrFlatJSON(data []byte, target any) error {
	if err := client.transport.paginator.DecodeSlurpedJSON(data, target); err == nil {
		return nil
	}
	return client.transport.decoder.DecodeJSON(data, target)
}

func (client *Client) runGH(_ string, args ...string) (CommandResult, error) {
	return client.execute(Command{Args: args, DisplayArgs: args})
}

func (client *Client) runGHWithInput(_ string, input []byte, args ...string) (CommandResult, error) {
	return client.execute(Command{Args: args, Stdin: input, DisplayArgs: args})
}
