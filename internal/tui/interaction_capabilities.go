package tui

type interactionCapabilitySnapshot struct {
	linkOpenerAvailable      bool
	clipboardReaderAvailable bool
}

func (program *Program) currentInteractionCapabilitySnapshot() interactionCapabilitySnapshot {
	if program == nil {
		return interactionCapabilitySnapshot{}
	}
	return interactionCapabilitySnapshot{
		linkOpenerAvailable:      program.linkOpener != nil,
		clipboardReaderAvailable: program.clipboardReader != nil,
	}
}
