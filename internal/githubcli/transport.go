package githubcli

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	appconfig "github.com/l-lin/lazygh/internal/config"
)

var ErrInvalidGraphQLResponse = errors.New("invalid GraphQL response")

type Command struct {
	Args        []string
	Stdin       []byte
	DisplayArgs []string
}

type Executor interface {
	Execute(Command) (CommandResult, error)
}

type CommandFormatter interface {
	Format(Command) string
}

type GraphQLClient interface {
	Query(GraphQLRequest) (CommandResult, error)
}

type RESTClient interface {
	Do(RESTRequest) (CommandResult, error)
}

type Paginator interface {
	DecodeSlurpedJSON(data []byte, target any) error
}

type ResponseDecoder interface {
	DecodeJSON(data []byte, target any) error
	DecodeGraphQL(data []byte, target any) error
}

type ErrorClassifier interface {
	Classify(Command, CommandResult, error) error
}

type GraphQLRequest struct {
	Query       string
	Variables   []GraphQLVariable
	DisplayArgs []string
}

type GraphQLVariable struct {
	Name  string
	Value string
	Typed bool
}

type RESTRequest struct {
	Path        string
	Method      string
	Headers     []RESTHeader
	Paginate    bool
	Slurp       bool
	Include     bool
	Input       []byte
	Flags       []string
	DisplayArgs []string
}

type RESTHeader struct {
	Name  string
	Value string
}

func literalGraphQLVariable(name string, value string) GraphQLVariable {
	return GraphQLVariable{Name: strings.TrimSpace(name), Value: value}
}

func typedGraphQLVariable(name string, value any) GraphQLVariable {
	return GraphQLVariable{Name: strings.TrimSpace(name), Value: fmt.Sprint(value), Typed: true}
}

type commandFormatter struct{}

type errorClassifier struct {
	formatter CommandFormatter
}

type commandExecutor struct {
	runner     Runner
	classifier ErrorClassifier
}

type graphQLClient struct {
	executor Executor
}

type restClient struct {
	executor Executor
}

type paginator struct{}

type responseDecoder struct{}

type sharedTransport struct {
	executor   Executor
	formatter  CommandFormatter
	graphql    GraphQLClient
	rest       RESTClient
	paginator  Paginator
	decoder    ResponseDecoder
	classifier ErrorClassifier
}

func NewCommandFormatter() CommandFormatter {
	return commandFormatter{}
}

func NewErrorClassifier(formatter CommandFormatter) ErrorClassifier {
	if formatter == nil {
		formatter = NewCommandFormatter()
	}
	return errorClassifier{formatter: formatter}
}

func NewExecutor(runner Runner, formatter CommandFormatter, classifier ErrorClassifier) Executor {
	if runner == nil {
		runner = execRunner{}
	}
	if formatter == nil {
		formatter = NewCommandFormatter()
	}
	if classifier == nil {
		classifier = NewErrorClassifier(formatter)
	}
	return commandExecutor{runner: runner, classifier: classifier}
}

func NewGraphQLClient(executor Executor) GraphQLClient {
	return graphQLClient{executor: executor}
}

func NewRESTClient(executor Executor) RESTClient {
	return restClient{executor: executor}
}

func NewPaginator() Paginator {
	return paginator{}
}

func NewResponseDecoder() ResponseDecoder {
	return responseDecoder{}
}

func newSharedTransport(runner Runner) sharedTransport {
	formatter := NewCommandFormatter()
	classifier := NewErrorClassifier(formatter)
	executor := NewExecutor(runner, formatter, classifier)
	return sharedTransport{
		executor:   executor,
		formatter:  formatter,
		graphql:    NewGraphQLClient(executor),
		rest:       NewRESTClient(executor),
		paginator:  NewPaginator(),
		decoder:    NewResponseDecoder(),
		classifier: classifier,
	}
}

func (formatter commandFormatter) Format(command Command) string {
	arguments := command.DisplayArgs
	if len(arguments) == 0 {
		arguments = command.Args
	}
	return appconfig.FormatGHCommand(arguments)
}

func (classifier errorClassifier) Classify(command Command, result CommandResult, err error) error {
	return classifyCommandError(classifier.formatter.Format(command), err, result.Stderr)
}

func (executor commandExecutor) Execute(command Command) (CommandResult, error) {
	if command.Stdin != nil {
		result, err := executor.runner.RunWithInput(ghBinaryName, command.Stdin, command.Args...)
		if err != nil {
			return CommandResult{}, executor.classifier.Classify(command, result, err)
		}
		return result, nil
	}

	result, err := executor.runner.Run(ghBinaryName, command.Args...)
	if err != nil {
		return CommandResult{}, executor.classifier.Classify(command, result, err)
	}
	return result, nil
}

func (client graphQLClient) Query(request GraphQLRequest) (CommandResult, error) {
	arguments := []string{"api", "graphql", "-f", "query=" + request.Query}
	for _, variable := range request.Variables {
		flag := "-f"
		if variable.Typed {
			flag = "-F"
		}
		arguments = append(arguments, flag, strings.TrimSpace(variable.Name)+"="+variable.Value)
	}

	displayArguments := request.DisplayArgs
	if len(displayArguments) == 0 {
		displayArguments = []string{"api", "graphql"}
	}

	return client.executor.Execute(Command{Args: arguments, DisplayArgs: displayArguments})
}

func (client restClient) Do(request RESTRequest) (CommandResult, error) {
	arguments := []string{"api", strings.TrimSpace(request.Path)}
	for _, header := range request.Headers {
		headerName := strings.TrimSpace(header.Name)
		headerValue := strings.TrimSpace(header.Value)
		if headerName == "" || headerValue == "" {
			continue
		}
		arguments = append(arguments, "-H", headerName+": "+headerValue)
	}
	if method := strings.TrimSpace(request.Method); method != "" {
		arguments = append(arguments, "--method", method)
	}
	arguments = append(arguments, request.Flags...)
	if request.Paginate {
		arguments = append(arguments, "--paginate")
	}
	if request.Slurp {
		arguments = append(arguments, "--slurp")
	}
	if request.Include {
		arguments = append(arguments, "--include")
	}
	if request.Input != nil {
		arguments = append(arguments, "--input", "-")
	}

	displayArguments := request.DisplayArgs
	if len(displayArguments) == 0 {
		displayArguments = append([]string(nil), arguments...)
	}

	return client.executor.Execute(Command{Args: arguments, Stdin: request.Input, DisplayArgs: displayArguments})
}

func (paginator) DecodeSlurpedJSON(data []byte, target any) error {
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Pointer || targetValue.IsNil() {
		return fmt.Errorf("decode slurped JSON: target must be a non-nil pointer")
	}
	if targetValue.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("decode slurped JSON: target must point to a slice")
	}

	sliceType := targetValue.Elem().Type()
	pagesType := reflect.SliceOf(sliceType)
	pagesPointer := reflect.New(pagesType)
	if err := json.Unmarshal(data, pagesPointer.Interface()); err != nil {
		return err
	}

	pages := pagesPointer.Elem()
	flattened := reflect.MakeSlice(sliceType, 0, 0)
	for pageIndex := 0; pageIndex < pages.Len(); pageIndex++ {
		flattened = reflect.AppendSlice(flattened, pages.Index(pageIndex))
	}
	targetValue.Elem().Set(flattened)
	return nil
}

func (responseDecoder) DecodeJSON(data []byte, target any) error {
	return json.Unmarshal(data, target)
}

func (responseDecoder) DecodeGraphQL(data []byte, target any) error {
	var envelope struct {
		Data   json.RawMessage `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return err
	}
	for _, graphqlErr := range envelope.Errors {
		message := strings.TrimSpace(graphqlErr.Message)
		if message != "" {
			return errors.New(message)
		}
	}
	if len(envelope.Data) == 0 || string(envelope.Data) == "null" {
		return ErrInvalidGraphQLResponse
	}
	return json.Unmarshal(envelope.Data, target)
}
