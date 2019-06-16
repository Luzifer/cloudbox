package providers

import (
	"github.com/pkg/errors"
)

type Capability uint8

const (
	CapBasic Capability = 1 << iota
	CapShare
	CapAutoChecksum
)

var (
	ErrInvalidURI          = errors.New("Spefified URI is invalid for this provider")
	ErrFeatureNotSupported = errors.New("Feature not supported")
)

type CloudProviderInitFunc func(string) (CloudProvider, error)

type CloudProvider interface {
	Capabilities() Capability
	Name() string
	DeleteFile(relativeName string) error
	GetFile(relativeName string) (File, error)
	ListFiles() ([]File, error)
	PutFile(File) error
	Share(relativeName string) (string, error)
}
