package handlers

import (
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/host"
)

func loadOneMachine(api libmachine.API, args ...string) (*host.Host, error) {
	if len(args) != 1 {
		return nil, errRequireMachineName
	}

	return api.Load(args[0])
}
