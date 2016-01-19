package main

import (
	"log"

	"github.com/dgageot/docker-machine-daemon/handlers"
	"github.com/dgageot/docker-machine-daemon/daemons"
)

const (
	sshPort = 2200
	httpPort = 8080
)

func main() {
	mappings := []handlers.Mapping{
		{"/machine/ls", handlers.RunLs},
	}

	go func() {
		log.Printf("Listening on %d...\n", sshPort)
		log.Printf(" - List the Docker Machines with: ssh localhost -p %d -s /machine/ls\n", sshPort)

		if err := daemons.NewSshDaemon(mappings).Start(sshPort); err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		log.Printf("Listening on %d...\n", httpPort)
		log.Printf(" - List the Docker Machines with: http GET http://localhost:%d/machine/ls\n", httpPort)

		if err := daemons.NewHttpDaemon(mappings).Start(httpPort); err != nil {
			log.Fatal(err)
		}
	}()

	select {}
}
