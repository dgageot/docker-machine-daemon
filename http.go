package main

import (
	"fmt"

	"github.com/docker/machine/libmachine"

	"net/http"

	"github.com/gorilla/mux"
)

func startHttpServer(port int, mappings []Mapping) error {
	r := mux.NewRouter()

	for _, mapping := range mappings {
		r.HandleFunc(mapping.url, toHandlerFunc(mapping.handler))
	}

	http.ListenAndServe(fmt.Sprintf(":%d", port), r)

	return nil
}

func toHandlerFunc(handler func(api libmachine.API) (interface{}, error)) http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		output, err := toJson(withApi(handler))
		if err != nil {
			response.WriteHeader(500)
			return
		}

		response.Write(output)
	}
}
