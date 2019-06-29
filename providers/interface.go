package providers

import (
	"hash"

	"github.com/pkg/errors"
)

type Capability uint8

const (
	CapBasic Capability = 1 << iota
	CapShare
	CapAutoChecksum
)

func (c Capability) Has(test Capability) bool { return c&test != 0 }

var (
	ErrInvalidURI          = errors.New("Spefified URI is invalid for this provider")
	ErrFeatureNotSupported = errors.New("Feature not supported")
)

type CloudProviderInitFunc func(string) (CloudProvider, error)

type CloudProvider interface {
	Capabilities() Capability
	DeleteFile(relativeName string) error
	GetChecksumMethod() hash.Hash
	GetFile(relativeName string) (File, error)
	ListFiles() ([]File, error)
	Name() string
	PutFile(File) (File, error)
	Share(relativeName string) (string, error)
}
