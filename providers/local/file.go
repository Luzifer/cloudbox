package local

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"os"

	"github.com/pkg/errors"

	"github.com/Luzifer/cloudbox/providers"
)

type File struct {
	info         os.FileInfo
	relativeName string
	fullPath     string
}

func (f File) Info() providers.FileInfo {
	return providers.FileInfo{
		RelativeName: f.relativeName,
		LastModified: f.info.ModTime(),
		Size:         uint64(f.info.Size()),
	}
}

func (f File) Checksum() (string, error) {
	fc, err := f.Content()
	if err != nil {
		return "", errors.Wrap(err, "Unable to get file contents")
	}

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, fc); err != nil {
		return "", errors.Wrap(err, "Unable to read file contents")
	}

	return fmt.Sprintf("%x", sha256.Sum256(buf.Bytes())), nil
}

func (f File) Content() (io.ReadCloser, error) {
	fp, err := os.Open(f.fullPath)
	return fp, errors.Wrap(err, "Unable to open file")
}
