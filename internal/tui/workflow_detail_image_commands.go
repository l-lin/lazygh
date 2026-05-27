package tui

import "github.com/jesseduffield/gocui"

type detailImageWorkflowRuntime struct {
	workflowShellRuntime
	renderMarkdownHTML func(string, string) (string, error)
	loadDetailImage    func(string) (loadedDetailImage, error)
}

type loadCurrentDetailImageHTMLCmd struct {
	source detailImageHTMLSource
}

type loadCurrentDetailImageCmd struct {
	imageURL string
}

func newDetailImageWorkflowRuntime(program *Program, gui *gocui.Gui) detailImageWorkflowRuntime {
	runtime := detailImageWorkflowRuntime{workflowShellRuntime: newWorkflowShellRuntime(program, gui)}
	if program == nil {
		return runtime
	}
	if program.markdownHTMLRenderer != nil {
		runtime.renderMarkdownHTML = program.markdownHTMLRenderer.RenderMarkdownHTML
	}
	runtime.loadDetailImage = func(imageURL string) (loadedDetailImage, error) {
		githubToken := ""
		if isGitHubImageSource(imageURL) {
			githubToken = program.detailImageAuthToken()
		}
		return loadDetailImage(imageURL, program.imageHTTPClient, githubToken)
	}
	return runtime
}

func (command loadCurrentDetailImageHTMLCmd) execute(program *Program, gui *gocui.Gui) {
	runtime := newDetailImageWorkflowRuntime(program, gui)
	if runtime.renderMarkdownHTML == nil || runtime.dispatchAsyncMessage == nil {
		return
	}
	source := command.source
	runWorkflowCommandAsync(runtime.runAsync, func() {
		renderedHTML, err := runtime.renderMarkdownHTML(source.repository, source.markdown)
		runtime.dispatchAsyncMessage(MsgCurrentDetailImageHTMLLoaded{Source: source, RenderedHTML: renderedHTML, Err: err})
	})
}

func (command loadCurrentDetailImageCmd) execute(program *Program, gui *gocui.Gui) {
	runtime := newDetailImageWorkflowRuntime(program, gui)
	if runtime.loadDetailImage == nil || runtime.dispatchAsyncMessage == nil {
		return
	}
	imageURL := command.imageURL
	runWorkflowCommandAsync(runtime.runAsync, func() {
		loadedImage, err := runtime.loadDetailImage(imageURL)
		runtime.dispatchAsyncMessage(MsgCurrentDetailImageLoaded{ImageURL: imageURL, Image: loadedImage, Err: err})
	})
}
