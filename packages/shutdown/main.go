package main

import (
	"github.com/openrelayxyz/plugeth-utils/core"
	"github.com/urfave/cli/v2"
)

type ShutdownService struct {
	stack   core.Node
}

var log core.Logger

func Initialize(ctx *cli.Context, loader core.PluginLoader, logger core.Logger) {
	log = logger
}

func GetAPIs(stack core.Node, backend core.Backend) []core.API {
	return []core.API{
		{
			Namespace: "admin",
			Version:   "1.0",
			Service:   &ShutdownService{stack},
		},
	}
}

func (service *ShutdownService) Shutdown() (bool, error) {
	return true, service.stack.Close()
}
