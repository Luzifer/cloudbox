package sync

import (
	"database/sql"
	"hash"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

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

	log *log.Entry

	stop chan struct{}
}

func New(local, remote providers.CloudProvider, db *sql.DB, conf SyncConfig, logger *log.Entry) *Sync {
	return &Sync{
		db:     db,
		conf:   conf,
		local:  local,
		remote: remote,

		log: logger,

		stop: make(chan struct{}),
	}
}

func (s *Sync) Run() error {
	if err := s.initSchema(); err != nil {
		return errors.Wrap(err, "Unable to initialize database schema")
	}

	var refresh = time.NewTimer(s.conf.ScanInterval)

	for {
		select {
		case <-refresh.C:
			if err := s.runSync(); err != nil {
				return errors.Wrap(err, "Sync failed")
			}
			refresh.Reset(s.conf.ScanInterval)

		case <-s.stop:
			return nil
		}
	}
}

func (s *Sync) Stop() { s.stop <- struct{}{} }

func (s *Sync) fillStateFromProvider(syncState *state, provider providers.CloudProvider, side string, useChecksum bool, hashMethod hash.Hash) error {
	files, err := provider.ListFiles()
	if err != nil {
		return errors.Wrap(err, "Unable to list files")
	}

	for _, f := range files {
		info := f.Info()
		if useChecksum && info.Checksum == "" {
			cs, err := f.Checksum(hashMethod)
			if err != nil {
				return errors.Wrap(err, "Unable to fetch checksum")
			}
			info.Checksum = cs
		}

		syncState.Set(side, sourceScan, info)
	}

	return nil
}

func (s *Sync) runSync() error {
	var (
		hashMethod  = s.remote.GetChecksumMethod()
		syncState   = newState()
		useChecksum = s.remote.Capabilities().Has(providers.CapAutoChecksum) || s.conf.ForceUseChecksum
	)

	if err := s.updateStateFromDatabase(syncState); err != nil {
		return errors.Wrap(err, "Unable to load database state")
	}

	if err := s.fillStateFromProvider(syncState, s.local, sideLocal, useChecksum, hashMethod); err != nil {
		return errors.Wrap(err, "Unable to load local files")
	}

	if err := s.fillStateFromProvider(syncState, s.remote, sideRemote, useChecksum, hashMethod); err != nil {
		return errors.Wrap(err, "Unable to load remote files")
	}

	// TODO: Do something with sync database
	s.log.Printf("%#v", syncState)

	return nil
}
