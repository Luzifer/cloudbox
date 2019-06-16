package main

import (
	"database/sql"
	"os"
	"os/signal"
	"path"
	"syscall"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/Luzifer/cloudbox/sync"
)

func execSync() error {
	conf, err := loadConfig(false)
	if err != nil {
		return errors.Wrap(err, "Unable to load config")
	}

	local, err := providerFromURI("file://" + conf.Sync.LocalDir)
	if err != nil {
		return errors.Wrap(err, "Unable to initialize local provider")
	}

	remote, err := providerFromURI(conf.Sync.RemoteURI)
	if err != nil {
		return errors.Wrap(err, "Unable to initialize remote provider")
	}

	if err := os.MkdirAll(conf.ControlDir, 0700); err != nil {
		return errors.Wrap(err, "Unable to create control dir")
	}

	db, err := sql.Open("sqlite3", path.Join(conf.ControlDir, "sync.db"))
	if err != nil {
		return errors.Wrap(err, "Unable to establish database connection")
	}

	s := sync.New(local, remote, db, conf.Sync.Settings)

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for range sigchan {
			s.Stop()
		}
	}()

	log.Info("Starting sync run...")
	return errors.Wrap(s.Run(), "Unable to sync")
}
