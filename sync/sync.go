package sync

import (
	"database/sql"

	"github.com/Luzifer/cloudbox/providers"
)

type Sync struct {
	db            *sql.DB
	local, remote providers.CloudProvider
}

func New(local, remote providers.CloudProvider, db *sql.DB) *Sync {
	return &Sync{
		db:     db,
		local:  local,
		remote: remote,
	}
}

func (s *Sync) Run() error {
	for {
		select {}
	}

	return nil
}
