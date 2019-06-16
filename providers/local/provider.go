package local

import (
	"crypto/sha256"
	"hash"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/Luzifer/cloudbox/providers"
)

func New(uri string) (providers.CloudProvider, error) {
	if !strings.HasPrefix(uri, "file://") {
		return nil, providers.ErrInvalidURI
	}

	return &Provider{directory: strings.TrimPrefix(uri, "file://")}, nil
}

type Provider struct {
	directory string
}

func (p Provider) Capabilities() providers.Capability { return providers.CapBasic }
func (p Provider) Name() string                       { return "local" }
func (p Provider) GetChecksumMethod() hash.Hash       { return sha256.New() }

func (p Provider) ListFiles() ([]providers.File, error) {
	return nil, errors.New("Not implemented")
}

func (p Provider) DeleteFile(relativeName string) error {
	return os.Remove(path.Join(p.directory, relativeName))
}

func (p Provider) GetFile(relativeName string) (providers.File, error) {
	fullPath := path.Join(p.directory, relativeName)

	stat, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, providers.ErrFileNotFound
		}
		return nil, errors.Wrap(err, "Unable to get file stat")
	}

	return File{
		info:         stat,
		relativeName: relativeName,
		fullPath:     fullPath,
	}, nil
}

func (p Provider) PutFile(f providers.File) (providers.File, error) {
	fullPath := path.Join(p.directory, f.Info().RelativeName)

	fp, err := os.Create(fullPath)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to create file")
	}

	rfp, err := f.Content()
	if err != nil {
		return nil, errors.Wrap(err, "Unable to get remote file content")
	}
	defer rfp.Close()

	if _, err := io.Copy(fp, rfp); err != nil {
		return nil, errors.Wrap(err, "Unable to copy file contents")
	}

	if err := fp.Close(); err != nil {
		return nil, errors.Wrap(err, "Unable to close local file")
	}

	if err := os.Chtimes(fullPath, time.Now(), f.Info().LastModified); err != nil {
		return nil, errors.Wrap(err, "Unable to set last file mod time")
	}

	return p.GetFile(f.Info().RelativeName)
}

func (p Provider) Share(relativeName string) (string, error) {
	return "", providers.ErrFeatureNotSupported
}
