package sync

import (
	"database/sql"
	"time"

	"github.com/pkg/errors"

	"github.com/Luzifer/cloudbox/providers"
)

type SyncConfig struct {
	ForceUseChecksum bool          `yaml:"force_use_checksum"`
	ScanInterval     time.Duration `yaml:"scan_interval"`
}

type Sync struct {
	db            *sql.DB
	conf          SyncConfig
	local, remote providers.CloudProvider

	stop chan struct{}
}

func New(local, remote providers.CloudProvider, db *sql.DB, conf SyncConfig) *Sync {
	return &Sync{
		db:     db,
		conf:   conf,
		local:  local,
		remote: remote,

		stop: make(chan struct{}),
	}
}

func (s *Sync) Run() error {
	if err := s.initSchema(); err != nil {
		return errors.Wrap(err, "Unable to initialize database schema")
	}

	var (
		hashMethod  = s.remote.GetChecksumMethod()
		refresh     = time.NewTimer(s.conf.ScanInterval)
		useChecksum = s.remote.Capabilities().Has(providers.CapAutoChecksum) || s.conf.ForceUseChecksum
	)

	for {
		select {
		case <-refresh.C:
			// TODO: Execute rescan & sync
			refresh.Reset(s.conf.ScanInterval)

		case <-s.stop:
			return nil
		}
	}
}

func (s *Sync) Stop() { s.stop <- struct{}{} }
