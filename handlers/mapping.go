package handlers

import (
	"encoding/json"

	"github.com/docker/machine/commands/mcndirs"
	"github.com/docker/machine/libmachine"
)

type Success struct {
	Action string
	Name   string
}

type Mapping struct {
	Url     string
	Handler Handler
}

func NewMapping(url string, handler HandlerFunc) Mapping {
	return Mapping{url, handler}
}

type Handler interface {
	Handle(api libmachine.API, args map[string]string) (interface{}, error)
}

type HandlerFunc func(api libmachine.API, args map[string]string) (interface{}, error)

func (f HandlerFunc) Handle(api libmachine.API, args map[string]string) (interface{}, error) {
	return f(api, args)
}

func WithApi(handler Handler, args map[string]string) func() (interface{}, error) {
	return func() (interface{}, error) {
		api := libmachine.NewClient(mcndirs.GetBaseDir(), mcndirs.GetMachineCertDir())
		defer api.Close()

		return handler.Handle(api, args)
	}
}

func ToJson(handler func() (interface{}, error)) ([]byte, error) {
	body, err := handler()
	if err != nil {
		return nil, err
	}

	return json.Marshal(body)
}
