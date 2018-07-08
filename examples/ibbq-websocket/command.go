package main

import (
	"github.com/containous/flaeg"
)

func newCommand(run func(*Configuration) error) *flaeg.Command {
	return &flaeg.Command{
		Name:                  "ibbq-websocket",
		Description:           "ibbq-websocket is a websocket server for ibbq devices",
		Config:                DefaultConfiguration,
		DefaultPointersConfig: &Configuration{},
		Run: func() error {
			return run(DefaultConfiguration)
		},
	}
}
