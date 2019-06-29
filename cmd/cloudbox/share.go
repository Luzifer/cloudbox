package main

import (
	"fmt"
	"os"
	"text/template"

	"github.com/pkg/errors"

	"github.com/Luzifer/cloudbox/providers"
	"github.com/Luzifer/rconfig"
)

func execShare() error {
	conf, err := loadConfig(false)
	if err != nil {
		return errors.Wrap(err, "Unable to load config")
	}

	remote, err := providerFromURI(conf.Sync.RemoteURI)
	if err != nil {
		return errors.Wrap(err, "Unable to initialize remote provider")
	}

	if !remote.Capabilities().Has(providers.CapShare) {
		return errors.New("Remote provider does not support sharing")
	}

	if len(rconfig.Args()) < 3 {
		return errors.New("No filename provided to share")
	}

	relativeName := rconfig.Args()[2]
	providerURL, err := remote.Share(relativeName)
	if err != nil {
		return errors.Wrap(err, "Unable to share file")
	}

	if !conf.Share.OverrideURI {
		fmt.Println(providerURL)
		return nil
	}

	tpl, err := template.New("share_uri").Parse(conf.Share.URITemplate)
	if err != nil {
		return errors.Wrap(err, "Unable to parse URI template")
	}

	if err := tpl.Execute(os.Stdout, map[string]interface{}{
		"file": relativeName,
	}); err != nil {
		return errors.Wrap(err, "Unable to render share URI")
	}

	return nil
}
