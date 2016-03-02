package handlers

import "github.com/docker/machine/libmachine"

// Restart restarts a Docker Machine
func Restart(api libmachine.API, args map[string]string) (interface{}, error) {
	h, err := loadOneMachine(api, args)
	if err != nil {
		return nil, err
	}

	if err := h.Restart(); err != nil {
		return nil, err
	}

	return Success{"restarted", h.Name}, nil
}
