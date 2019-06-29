package providers

import (
	"hash"
	"io"
	"time"

	"github.com/pkg/errors"
)

var ErrFileNotFound = errors.New("File not found")

type File interface {
	Info() FileInfo
	Checksum(hash.Hash) (string, error)
	Content() (io.ReadCloser, error)
}

type FileInfo struct {
	RelativeName string
	LastModified time.Time
	Checksum     string // Expected to be present on CapAutoChecksum
	Size         uint64
}

func (f *FileInfo) Equal(other *FileInfo) bool {
	if f == nil && other == nil {
		// Both are not present: No change
		return true
	}

	if (f != nil && other == nil) || (f == nil && other != nil) {
		// One is not present, the other is: Change
		return false
	}

	if (f.Checksum != "" || other.Checksum != "") && f.Checksum != other.Checksum {
		// Checksum is present in one, doesn't match: Change
		return false
	}

	if f.Size != other.Size {
		// No checksums present, size differs: Change
		return false
	}

	if !f.LastModified.Equal(other.LastModified) {
		// LastModified date differs: Change
		return false
	}

	// No changes detected yet: No change
	return true
}
