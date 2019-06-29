package sync

import (
	"database/sql"
	"hash"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/Luzifer/cloudbox/providers"
)

type Config struct {
	ForceUseChecksum bool          `yaml:"force_use_checksum"`
	ScanInterval     time.Duration `yaml:"scan_interval"`
}

type Sync struct {
	db            *sql.DB
	conf          Config
	local, remote providers.CloudProvider

	log *log.Entry

	useChecksum bool
	hashMethod  hash.Hash

	stop chan struct{}
}

func New(local, remote providers.CloudProvider, db *sql.DB, conf Config, logger *log.Entry) *Sync {
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

func (s *Sync) getFileInfo(f providers.File) (providers.FileInfo, error) {
	var info = f.Info()

	if !s.useChecksum || info.Checksum != "" {
		return info, nil
	}

	cs, err := f.Checksum(s.hashMethod)
	if err != nil {
		return info, errors.Wrap(err, "Unable to fetch checksum")
	}
	info.Checksum = cs

	return info, nil
}

func (s *Sync) fillStateFromProvider(syncState *state, provider providers.CloudProvider, side string) error {
	files, err := provider.ListFiles()
	if err != nil {
		return errors.Wrap(err, "Unable to list files")
	}

	for _, f := range files {
		info, err := s.getFileInfo(f)
		if err != nil {
			return errors.Wrap(err, "Unable to get file info")
		}

		syncState.Set(side, sourceScan, info)
	}

	return nil
}

func (s *Sync) runSync() error {
	var syncState = newState()
	s.hashMethod = s.remote.GetChecksumMethod()
	s.useChecksum = s.remote.Capabilities().Has(providers.CapAutoChecksum) || s.conf.ForceUseChecksum

	if err := s.updateStateFromDatabase(syncState); err != nil {
		return errors.Wrap(err, "Unable to load database state")
	}

	if err := s.fillStateFromProvider(syncState, s.local, sideLocal); err != nil {
		return errors.Wrap(err, "Unable to load local files")
	}

	if err := s.fillStateFromProvider(syncState, s.remote, sideRemote); err != nil {
		return errors.Wrap(err, "Unable to load remote files")
	}

	for _, fileName := range syncState.GetRelativeNames() {
		if err := s.decideAction(syncState, fileName); err != nil {
			return errors.Wrap(err, "Could not execute sync")
		}
	}

	return nil
}
