package clipboard

import systemclipboard "github.com/atotto/clipboard"

type Writer interface {
	WriteText(text string) error
}

type SystemWriter struct {
	writeAll func(string) error
}

func NewSystemWriter() *SystemWriter {
	return &SystemWriter{writeAll: systemclipboard.WriteAll}
}

func (writer SystemWriter) WriteText(text string) error {
	writeAll := writer.writeAll
	if writeAll == nil {
		writeAll = systemclipboard.WriteAll
	}

	return writeAll(text)
}
