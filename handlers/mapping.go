package handlers

import (
	"encoding/json"

	"sync"

	"github.com/docker/machine/commands/mcndirs"
	"github.com/docker/machine/libmachine"
)

// libmachine is not thread safe, specially when it saves machines to the disk
var globalMutex = &sync.Mutex{}

type Success struct {
	Action string
	Name   string
}

type Mapping struct {
	Method  string
	Url     string
	Handler Handler
}

func NewMapping(method string, url string, handler HandlerFunc) Mapping {
	return Mapping{method, url, handler}
}

type Handler interface {
	Handle(api libmachine.API, args map[string]string, form map[string][]string) (interface{}, error)
}

type HandlerFunc func(api libmachine.API, args map[string]string, form map[string][]string) (interface{}, error)

func (f HandlerFunc) Handle(api libmachine.API, args map[string]string, form map[string][]string) (interface{}, error) {
	return f(api, args, form)
}

func WithApi(handler Handler, args map[string]string, form map[string][]string) func() (interface{}, error) {
	return func() (interface{}, error) {
		globalMutex.Lock()
		defer globalMutex.Unlock()

		api := libmachine.NewClient(mcndirs.GetBaseDir(), mcndirs.GetMachineCertDir())
		defer api.Close()

		return handler.Handle(api, args, form)
	}
}

func ToJson(handler func() (interface{}, error)) ([]byte, error) {
	body, err := handler()
	if err != nil {
		return nil, err
	}

	return json.Marshal(body)
}
