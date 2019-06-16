package main

import (
	"github.com/pkg/errors"

	"github.com/Luzifer/cloudbox/providers"
	"github.com/Luzifer/cloudbox/providers/local"
	"github.com/Luzifer/cloudbox/providers/s3"
)

var providerInitFuncs = []providers.CloudProviderInitFunc{
	local.New,
	s3.New,
}

func providerFromURI(uri string) (providers.CloudProvider, error) {
	if uri == "" {
		return nil, errors.New("Empty provider URI")
	}

	for _, f := range providerInitFuncs {
		cp, err := f(uri)
		switch err {
		case nil:
			if !cp.Capabilities().Has(providers.CapBasic) {
				return nil, errors.Errorf("Provider %s does not support basic capabilities", cp.Name())
			}

			return cp, nil
		case providers.ErrInvalidURI:
			// Fine for now, try next one
		default:
			return nil, errors.Wrap(err, "Unable to initialize provider")
		}
	}

	return nil, errors.Errorf("No provider found for URI %q", uri)
}
