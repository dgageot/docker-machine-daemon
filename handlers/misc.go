package handlers

import (
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/commands/mcndirs"
	"encoding/json"
)

func WithApi(handler func(api libmachine.API) (interface{}, error)) func() (interface{}, error) {
	return func() (interface{}, error) {
		api := libmachine.NewClient(mcndirs.GetBaseDir(), mcndirs.GetMachineCertDir())
		defer api.Close()

		return handler(api)
	}
}

func ToJson(handler func() (interface{}, error)) ([]byte, error) {
	body, err := handler()
	if err != nil {
		return nil, err
	}

	return json.Marshal(body)
}
