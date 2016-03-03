package handlers

import (
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/mcnerror"
)

// Stop stops a Docker Machine
func Stop(api libmachine.API, args map[string]string, form map[string][]string) (interface{}, error) {
	h, err := loadOneMachine(api, args)
	if err != nil {
		return nil, err
	}

	if err := h.Stop(); err != nil {
		if _, alreadyStated := err.(mcnerror.ErrHostAlreadyInState); alreadyStated {
			return nil, err
		}
	}

	return Success{"stopped", h.Name}, nil
}
