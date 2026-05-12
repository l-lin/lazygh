package githubcli

type serviceBase struct {
	transport sharedTransport
}

func newServiceBase(transport sharedTransport) serviceBase {
	return serviceBase{transport: transport}
}

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

func (service serviceBase) execute(command Command) (CommandResult, error) {
	return service.transport.executor.Execute(command)
}

func (service serviceBase) queryGraphQL(request GraphQLRequest) (CommandResult, error) {
	return service.transport.graphql.Query(request)
}

func (service serviceBase) doREST(request RESTRequest) (CommandResult, error) {
	return service.transport.rest.Do(request)
}

func (service serviceBase) decodePaginatedOrFlatJSON(data []byte, target any) error {
	if err := service.transport.paginator.DecodeSlurpedJSON(data, target); err == nil {
		return nil
	}
	return service.transport.decoder.DecodeJSON(data, target)
}

func (client *Client) runGH(_ string, args ...string) (CommandResult, error) {
	return client.execute(Command{Args: args, DisplayArgs: args})
}

func (client *Client) runGHWithInput(_ string, input []byte, args ...string) (CommandResult, error) {
	return client.execute(Command{Args: args, Stdin: input, DisplayArgs: args})
}
