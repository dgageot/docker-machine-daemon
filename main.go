package main

import (
	"log"

	"github.com/dgageot/docker-machine-daemon/daemon/http"
	"github.com/dgageot/docker-machine-daemon/handlers"
)

const (
	httpPort = 8080
)

func main() {
	mappings := []handlers.Mapping{
		handlers.NewMappingFunc("/machine/ls", handlers.Ls),
		handlers.NewMappingFunc("/machine/{machine}/start", handlers.Start),
		handlers.NewMappingFunc("/machine/{machine}/stop", handlers.Stop),
		handlers.NewMappingFunc("/machine/{machine}/restart", handlers.Restart),
	}
	daemon := http.NewDaemon(mappings)

	log.Printf("Listening on %d...\n", httpPort)
	log.Printf(" - List the Docker Machines with: http GET http://localhost:%d/machine/ls\n", httpPort)

	if err := daemon.Start(httpPort); err != nil {
		log.Fatal(err)
	}
}
