package main

import (
	"log"

	"github.com/dgageot/docker-machine-daemon/daemon/http"
	"github.com/dgageot/docker-machine-daemon/daemon/ssh"
	"github.com/dgageot/docker-machine-daemon/handlers"
)

const (
	sshPort  = 2200
	httpPort = 8080
)

func main() {
	go func() {
		mappings := []handlers.Mapping{
			handlers.NewMappingFunc("/machine/ls", handlers.Ls),
			handlers.NewMappingFunc("/machine/start", handlers.Start),
			handlers.NewMappingFunc("/machine/stop", handlers.Stop),
			handlers.NewMappingFunc("/machine/restart", handlers.Restart),
		}
		daemon := ssh.NewDaemon(mappings)

		log.Printf("Listening on %d...\n", sshPort)
		log.Printf(" - List the Docker Machines with: ssh localhost -p %d -s /machine/ls\n", sshPort)

		if err := daemon.Start(sshPort); err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
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
	}()

	select {}
}
