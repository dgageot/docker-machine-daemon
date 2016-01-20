package handlers

import "github.com/docker/machine/libmachine"

// Start starts a Docker Machine
func Start(api libmachine.API, args ...string) (interface{}, error) {
	h, err := loadOneMachine(api, args...)
	if err != nil {
		return nil, err
	}

	if err := h.Start(); err != nil {
		return nil, err
	}

	return Success{"started", h.Name}, nil
}
