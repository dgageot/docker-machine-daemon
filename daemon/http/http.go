package http

import (
	"fmt"

	"net/http"

	"log"

	"github.com/dgageot/docker-machine-daemon/daemon"
	"github.com/dgageot/docker-machine-daemon/handlers"
	"github.com/gorilla/mux"
)

type httpDaemon struct {
	mappings []handlers.Mapping
}

// NewDaemon create a new http daemon with given mappings.
func NewDaemon(mappings ...handlers.Mapping) daemon.Starter {
	return &httpDaemon{
		mappings: mappings,
	}
}

// Start starts the http daemon.
func (d *httpDaemon) Start(port int) error {
	r := mux.NewRouter()

	for _, mapping := range d.mappings {
		r.NewRoute().Path(mapping.Url).Handler(toHandler(mapping.Handler)).Methods(mapping.Method)
	}

	return http.ListenAndServe(fmt.Sprintf(":%d", port), r)
}

func toHandler(handler handlers.Handler) http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		if err := request.ParseForm(); err != nil {
			log.Print(err)
			response.WriteHeader(500)
			return
		}

		output, err := handlers.ToJson(handlers.WithApi(handler, mux.Vars(request), request.PostForm))
		if err != nil {
			log.Print(err)
			response.WriteHeader(500)
			return
		}

		response.Header().Set("Content-Type", "application/json")
		response.Write(output)
	}
}
