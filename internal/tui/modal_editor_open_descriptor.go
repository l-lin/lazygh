package tui

type modalEditorOpenDescriptor struct {
	kind             modalEditorKind
	title            string
	initialText      string
	submitDescriptor modalEditorSubmitDescriptor
	totalHeight      int
}

func newModalEditorOpenDescriptor(title string, initialText string) modalEditorOpenDescriptor {
	return newMultilineModalEditorOpenDescriptor(title, initialText, modalEditorTotalHeight)
}

func newModalEditorOpenDescriptorWithSubmitDescriptor(title string, initialText string, submitDescriptor modalEditorSubmitDescriptor) modalEditorOpenDescriptor {
	descriptor := newModalEditorOpenDescriptor(title, initialText)
	descriptor.submitDescriptor = submitDescriptor
	return descriptor
}

func newLineModalEditorOpenDescriptor(title string, initialText string) modalEditorOpenDescriptor {
	return newLineModalEditorOpenDescriptorWithHeight(title, initialText, lineModalEditorTotalHeight)
}

func newLineModalEditorOpenDescriptorWithSubmitDescriptor(title string, initialText string, submitDescriptor modalEditorSubmitDescriptor) modalEditorOpenDescriptor {
	return newLineModalEditorOpenDescriptorWithHeightAndSubmitDescriptor(title, initialText, submitDescriptor, lineModalEditorTotalHeight)
}

func newLineModalEditorOpenDescriptorWithHeight(title string, initialText string, totalHeight int) modalEditorOpenDescriptor {
	if totalHeight < 1 {
		totalHeight = lineModalEditorTotalHeight
	}
	return modalEditorOpenDescriptor{
		kind:        modalEditorKindSingleLine,
		title:       title,
		initialText: initialText,
		totalHeight: totalHeight,
	}
}

func newLineModalEditorOpenDescriptorWithHeightAndSubmitDescriptor(title string, initialText string, submitDescriptor modalEditorSubmitDescriptor, totalHeight int) modalEditorOpenDescriptor {
	descriptor := newLineModalEditorOpenDescriptorWithHeight(title, initialText, totalHeight)
	descriptor.submitDescriptor = submitDescriptor
	return descriptor
}

func newMultilineModalEditorOpenDescriptor(title string, initialText string, totalHeight int) modalEditorOpenDescriptor {
	return modalEditorOpenDescriptor{
		kind:        modalEditorKindMultiline,
		title:       title,
		initialText: initialText,
		totalHeight: totalHeight,
	}
}

func newMultilineModalEditorOpenDescriptorWithSubmitDescriptor(title string, initialText string, submitDescriptor modalEditorSubmitDescriptor, totalHeight int) modalEditorOpenDescriptor {
	descriptor := newMultilineModalEditorOpenDescriptor(title, initialText, totalHeight)
	descriptor.submitDescriptor = submitDescriptor
	return descriptor
}

func (descriptor modalEditorOpenDescriptor) state() modalEditorState {
	switch descriptor.kind {
	case modalEditorKindSingleLine:
		return newLineModalEditorStateWithHeightAndSubmitDescriptor(descriptor.title, descriptor.initialText, descriptor.submitDescriptor, descriptor.totalHeight)
	case modalEditorKindMultiline:
		return newMultilineModalEditorStateWithSubmitDescriptor(descriptor.title, descriptor.initialText, descriptor.submitDescriptor, descriptor.totalHeight)
	default:
		return modalEditorState{}
	}
}
