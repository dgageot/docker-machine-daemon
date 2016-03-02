package handlers

import (
	"strings"

	"github.com/docker/machine/libmachine"
)

// Stop stops a Docker Machine
func Stop(api libmachine.API, args map[string]string, form map[string][]string) (interface{}, error) {
	h, err := loadOneMachine(api, args)
	if err != nil {
		return nil, err
	}

	if err := h.Stop(); err != nil {
		// TODO: machine should return a type error
		if !strings.Contains(err.Error(), "is already stopped") {
			return nil, err
		}
	}

	return Success{"stopped", h.Name}, nil
}
