package tui

import (
	"github.com/jesseduffield/gocui"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

type connectedUserWorkflowRuntime struct {
	workflowShellRuntime
	getConnectedUser func() (githubdomain.ConnectedUser, error)
}

type loadConnectedUserCmd struct{}

func newConnectedUserWorkflowRuntime(program *Program, gui *gocui.Gui) connectedUserWorkflowRuntime {
	runtime := connectedUserWorkflowRuntime{workflowShellRuntime: newWorkflowShellRuntime(program, gui)}
	if program != nil && program.sessionQueries != nil {
		runtime.getConnectedUser = program.sessionQueries.GetConnectedUser
	}
	return runtime
}

func (loadConnectedUserCmd) execute(program *Program, gui *gocui.Gui) {
	runtime := newConnectedUserWorkflowRuntime(program, gui)
	if runtime.getConnectedUser == nil || runtime.dispatchAsyncMessage == nil {
		return
	}
	runWorkflowCommandAsync(runtime.runAsync, func() {
		user, err := runtime.getConnectedUser()
		runtime.dispatchAsyncMessage(MsgConnectedUserLoaded{User: user, Err: err})
	})
}
