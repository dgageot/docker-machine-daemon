package handlers

import (
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/mcnerror"
)

// Start starts a Docker Machine
func Start(api libmachine.API, args map[string]string, form map[string][]string) (interface{}, error) {
	h, err := loadOneMachine(api, args)
	if err != nil {
		return nil, err
	}

	if err := h.Start(); err != nil {
		if _, alreadyStated := err.(mcnerror.ErrHostAlreadyInState); alreadyStated {
			return nil, err
		}
	}

	return Success{"started", h.Name}, nil
}
