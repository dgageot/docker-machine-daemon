package handlers

import "errors"

var (
	errRequireMachineName = errors.New("Requires one machine name")
	errRequireDriverName  = errors.New("Requires a driver name")
)
