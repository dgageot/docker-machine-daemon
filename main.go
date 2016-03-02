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
	daemon := http.NewDaemon(
		handlers.NewMapping("GET", "/machine/ls", handlers.Ls),
		handlers.NewMapping("POST", "/machine/{name}/start", handlers.Start),
		handlers.NewMapping("POST", "/machine/{name}/stop", handlers.Stop),
		handlers.NewMapping("POST", "/machine/{name}/restart", handlers.Restart),
	)

	log.Printf("Listening on %d...\n", httpPort)
	log.Printf(" - List the Docker Machines with: http GET http://localhost:%d/machine/ls\n", httpPort)

	if err := daemon.Start(httpPort); err != nil {
		log.Fatal(err)
	}
}
