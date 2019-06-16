package main

import (
	"os"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/Luzifer/cloudbox/sync"
)

type shareConfig struct {
	OverrideURI bool   `yaml:"override_uri"`
	URITemplate string `yaml:"uri_template"`
}

type syncConfig struct {
	LocalDir  string          `yaml:"local_dir"`
	RemoteURI string          `yaml:"remote_uri"`
	Settings  sync.SyncConfig `yaml:"settings"`
}

type configFile struct {
	ControlDir string      `yaml:"control_dir"`
	Sync       syncConfig  `yaml:"sync"`
	Share      shareConfig `yaml:"share"`
}

func (c configFile) validate() error {
	if c.Sync.LocalDir == "" {
		return errors.New("Local directory not specified")
	}

	if c.Sync.RemoteURI == "" {
		return errors.New("Remote sync URI not specified")
	}

	if c.Share.OverrideURI && c.Share.URITemplate == "" {
		return errors.New("Share URI override enabled but no template specified")
	}

	return nil
}

func defaultConfig() *configFile {
	return &configFile{
		ControlDir: "~/.cache/cloudbox",
		Sync: syncConfig{
			Settings: sync.SyncConfig{
				ScanInterval: time.Minute,
			},
		},
	}
}

func execWriteSampleConfig() error {
	conf := defaultConfig()

	if _, err := os.Stat(cfg.Config); err == nil {
		if conf, err = loadConfig(true); err != nil {
			return errors.Wrap(err, "Unable to load existing config")
		}
	}

	f, err := os.Create(cfg.Config)
	if err != nil {
		return errors.Wrap(err, "Unable to create config file")
	}
	defer f.Close()

	f.WriteString("---\n\n")

	if err := yaml.NewEncoder(f).Encode(conf); err != nil {
		return errors.Wrap(err, "Unable to write config file")
	}

	f.WriteString("\n...\n")

	log.WithField("dest", cfg.Config).Info("Config written")
	return nil
}

func loadConfig(noValidate bool) (*configFile, error) {
	config := defaultConfig()

	f, err := os.Open(cfg.Config)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to open config file")
	}
	defer f.Close()

	if err = yaml.NewDecoder(f).Decode(config); err != nil {
		return nil, errors.Wrap(err, "Unable to decode config")
	}

	if noValidate {
		return config, nil
	}

	return config, config.validate()
}
