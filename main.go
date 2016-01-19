package main

import (
	"encoding/json"
	"log"

	"github.com/dgageot/docker-machine-daemon/ls"
	"github.com/docker/machine/commands/mcndirs"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/persist"
)

func main() {
	go func() {
		if err := startSshDaemon(); err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		if err := startHttpServer(); err != nil {
			log.Fatal(err)
		}
	}()

	select {}
}

// runLs lists all machines.
func runLs(api libmachine.API) (interface{}, error) {
	hostList, hostInError, err := persist.LoadAllHosts(api)
	if err != nil {
		return nil, err
	}

	return ls.GetHostListItems(hostList, hostInError), nil
}

func withApi(handler func(api libmachine.API) (interface{}, error)) func() (interface{}, error) {
	return func() (interface{}, error) {
		api := libmachine.NewClient(mcndirs.GetBaseDir(), mcndirs.GetMachineCertDir())
		defer api.Close()

		return handler(api)
	}
}

func toJson(handler func() (interface{}, error)) ([]byte, error) {
	body, err := handler()
	if err != nil {
		return nil, err
	}

	return json.Marshal(body)
}
