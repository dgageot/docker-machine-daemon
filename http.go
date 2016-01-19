package main

import (
	"fmt"
	"log"

	"github.com/docker/machine/libmachine"

	"net/http"

	"github.com/gorilla/mux"
)

const (
	httpPort = 8080
)

func startHttpServer() error {
	log.Printf("Listening on %d...\n", httpPort)
	log.Printf(" - List the Docker Machines with: http GET http://localhost:%d/machine/ls\n", httpPort)

	r := mux.NewRouter()
	r.HandleFunc("/machine/ls", toHandlerFunc(runLs))

	http.ListenAndServe(fmt.Sprintf(":%d", httpPort), r)

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
