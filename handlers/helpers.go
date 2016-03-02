package handlers

import (
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/host"
)

func loadOneMachine(api libmachine.API, args map[string]string) (*host.Host, error) {
	name, present := args["name"]
	if !present {
		return nil, errRequireMachineName
	}

	return api.Load(name)
}
