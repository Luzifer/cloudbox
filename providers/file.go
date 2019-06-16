package providers

import (
	"io"
	"time"

	"github.com/pkg/errors"
)

var ErrFileNotFound = errors.New("File not found")

type File interface {
	Info() FileInfo
	Checksum() (string, error)
	Content() (io.ReadCloser, error)
}

type FileInfo struct {
	RelativeName string
	LastModified time.Time
	Checksum     string // Expected to be present on CapAutoChecksum
	Size         uint64
}
