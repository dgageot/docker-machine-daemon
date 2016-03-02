package handlers

import (
	"strings"

	"github.com/docker/machine/libmachine"
)

// Start starts a Docker Machine
func Start(api libmachine.API, args map[string]string) (interface{}, error) {
	h, err := loadOneMachine(api, args)
	if err != nil {
		return nil, err
	}

	if err := h.Start(); err != nil {
		// TODO: machine should return a type error
		if !strings.Contains(err.Error(), "is already running") {
			return nil, err
		}
	}

	return Success{"started", h.Name}, nil
}
