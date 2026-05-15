package clipboard

import systemclipboard "github.com/atotto/clipboard"

type Reader interface {
	ReadText() (string, error)
}

type SystemReader struct {
	readAll func() (string, error)
}

func NewSystemReader() *SystemReader {
	return &SystemReader{readAll: systemclipboard.ReadAll}
}

func (reader SystemReader) ReadText() (string, error) {
	readAll := reader.readAll
	if readAll == nil {
		readAll = systemclipboard.ReadAll
	}

	return readAll()
}
