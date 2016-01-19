package handlers

import "github.com/docker/machine/libmachine"

type Mapping struct {
	Url     string
	Handler func(api libmachine.API) (interface{}, error)
}
