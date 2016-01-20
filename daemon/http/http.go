package http

import (
	"fmt"

	"net/http"

	"github.com/dgageot/docker-machine-daemon/daemon"
	"github.com/dgageot/docker-machine-daemon/handlers"
	"github.com/gorilla/mux"
)

type httpDaemon struct {
	mappings []handlers.Mapping
}

// NewDaemon create a new http daemon with given mappings.
func NewDaemon(mappings []handlers.Mapping) daemon.Starter {
	return &httpDaemon{
		mappings: mappings,
	}
}

// Start starts the http daemon.
func (d *httpDaemon) Start(port int) error {
	r := mux.NewRouter()

	for _, mapping := range d.mappings {
		r.Handle(mapping.Url, toHandler(mapping.Handler))
	}

	return http.ListenAndServe(fmt.Sprintf(":%d", port), r)
}

func toHandler(handler handlers.Handler) http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		output, err := handlers.ToJson(handlers.WithApi(handler))
		if err != nil {
			response.WriteHeader(500)
		} else {
			response.Header().Set("Content-Type", "application/json")
			response.Write(output)
		}
	}
}
