package daemons

import (
	"fmt"

	"github.com/docker/machine/libmachine"

	"net/http"

	"github.com/dgageot/docker-machine-daemon/handlers"
	"github.com/gorilla/mux"
)

type httpDaemon struct {
	mappings []handlers.Mapping
}

// NewHttpDaemon create a new http daemon with given mappings.
func NewHttpDaemon(mappings []handlers.Mapping) Starter {
	return &httpDaemon{
		mappings: mappings,
	}
}

// Start startsth http daemon.
func (d *httpDaemon) Start(port int) error {
	r := mux.NewRouter()

	for _, mapping := range d.mappings {
		r.HandleFunc(mapping.Url, toHandlerFunc(mapping.Handler))
	}

	return http.ListenAndServe(fmt.Sprintf(":%d", port), r)
}

func toHandlerFunc(handler func(api libmachine.API) (interface{}, error)) http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		output, err := handlers.ToJson(handlers.WithApi(handler))
		if err != nil {
			response.WriteHeader(500)
		} else {
			response.Write(output)
		}
	}
}
