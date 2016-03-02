package handlers

import "github.com/docker/machine/libmachine"

// Remove removes a Docker Machine
func Remove(api libmachine.API, args map[string]string, form map[string][]string) (interface{}, error) {
	name, present := args["name"]
	if !present {
		return nil, errRequireMachineName
	}

	exist, _ := api.Exists(name)
	if !exist {
		return Success{"removed", name}, nil
	}

	currentHost, err := api.Load(name)
	if err != nil {
		return nil, err
	}

	if err := currentHost.Driver.Remove(); err != nil {
		return nil, err
	}

	if err := api.Remove(name); err != nil {
		return nil, err
	}

	return Success{"removed", name}, nil
}
