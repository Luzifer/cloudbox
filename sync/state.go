package sync

import (
	"sort"
	"strings"
	"sync"

	"github.com/Luzifer/cloudbox/providers"
)

type Change uint8

const (
	ChangeLocalAdd Change = 1 << iota
	ChangeLocalDelete
	ChangeLocalUpdate
	ChangeRemoteAdd
	ChangeRemoteDelete
	ChangeRemoteUpdate
)

func (c Change) Changed() bool {
	return c != 0
}

func (c *Change) Register(add Change) {
	*c = *c | add
}

func (c Change) Has(test Change) bool {
	return c&test != 0
}

func (c Change) Is(test Change) bool {
	return c == test
}

const (
	SideLocal  string = "local"
	SideRemote string = "remote"
	SourceDB   string = "db"
	SourceScan string = "scan"
)

type stateDetail struct {
	LocalDB,
	LocalScan,
	RemoteDB,
	RemoteScan *providers.FileInfo
}

type state struct {
	files map[string]*stateDetail
	lock  sync.Mutex
}

func newState() *state {
	return &state{
		files: make(map[string]*stateDetail),
	}
}

func (s *state) GetChangeFor(relativeName string) (result Change) {
	s.lock.Lock()
	defer s.lock.Unlock()

	d := s.files[relativeName]

	// No changes detected
	if d.LocalDB.Equal(d.LocalScan) && d.RemoteDB.Equal(d.RemoteScan) {
		// Check special case: Something went really wrong and sync state is FUBAR
		if d.LocalDB == nil && d.RemoteDB != nil {
			result.Register(ChangeRemoteAdd)
		}
		if d.LocalDB != nil && d.RemoteDB == nil {
			result.Register(ChangeLocalAdd)
		}

		return
	}

	// Check for local changes
	switch {
	case d.LocalDB == nil && d.LocalScan != nil:
		result.Register(ChangeLocalAdd)

	case d.LocalDB != nil && d.LocalScan == nil:
		result.Register(ChangeLocalDelete)

	case !d.LocalDB.Equal(d.LocalScan):
		result.Register(ChangeLocalUpdate)
	}

	// Check for remote changes
	switch {
	case d.RemoteDB == nil && d.RemoteScan != nil:
		result.Register(ChangeRemoteAdd)

	case d.RemoteDB != nil && d.RemoteScan == nil:
		result.Register(ChangeRemoteDelete)

	case !d.RemoteDB.Equal(d.RemoteScan):
		result.Register(ChangeRemoteUpdate)
	}

	return
}

func (s *state) GetRelativeNames() []string {
	s.lock.Lock()
	defer s.lock.Unlock()

	out := []string{}
	for k := range s.files {
		out = append(out, k)
	}

	sort.Strings(out)

	return out
}

func (s *state) Set(side, source, relativeName string, info providers.FileInfo) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if _, ok := s.files[relativeName]; !ok {
		s.files[relativeName] = &stateDetail{}
	}

	switch strings.Join([]string{side, source}, "::") {
	case "local::db":
		s.files[relativeName].LocalDB = &info
	case "local::scan":
		s.files[relativeName].LocalScan = &info
	case "remote::db":
		s.files[relativeName].RemoteDB = &info
	case "remote::scan":
		s.files[relativeName].RemoteScan = &info
	}
}