package main

import (
	"fmt"

	"github.com/Luzifer/rconfig"
)

const helpText = `
Available commands:
  help            Display this message
  sync            Executes the bi-directional sync
  write-config    Write a sample configuration to specified location
`

func execHelp() error {
	rconfig.Usage()

	fmt.Print(helpText)
	return nil
}
