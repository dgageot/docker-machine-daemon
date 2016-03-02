package handlers

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/persist"

	"github.com/docker/machine/commands"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/mcndockerclient"
	"github.com/docker/machine/libmachine/state"
	"github.com/docker/machine/libmachine/swarm"
)

const (
	lsTimeoutDuration = 10 * time.Second
)

// Ls lists all Docker Machines.
func Ls(api libmachine.API, args map[string]string) (interface{}, error) {
	hostList, hostInError, err := persist.LoadAllHosts(api)
	if err != nil {
		return nil, err
	}

	return listHosts(hostList, hostInError), nil
}

func listHosts(validHosts []*host.Host, hostsInError map[string]error) []commands.HostListItem {
	itemChan := make(chan commands.HostListItem)
	for _, h := range validHosts {
		go getHostItem(h, itemChan)
	}

	hosts := []commands.HostListItem{}
	for range validHosts {
		hosts = append(hosts, <-itemChan)
	}

	close(itemChan)

	for name, err := range hostsInError {
		hosts = append(hosts, commands.HostListItem{
			Name:       name,
			DriverName: "not found",
			State:      state.Error,
			Error:      err.Error(),
		})
	}

	return hosts
}

func getHostItem(h *host.Host, itemChan chan<- commands.HostListItem) {
	hosts := make(chan commands.HostListItem)

	go attemptGetHostItem(h, hosts)

	select {
	case hli := <-hosts:
		itemChan <- hli
	case <-time.After(lsTimeoutDuration):
		itemChan <- commands.HostListItem{
			Name:       h.Name,
			DriverName: h.Driver.DriverName(),
			State:      state.Timeout,
		}
	}
}

// PERFORMANCE: The code of this function is complicated because we try
// to call the underlying drivers as less as possible to get the information
// we need.
func attemptGetHostItem(h *host.Host, stateQueryChan chan<- commands.HostListItem) {
	url := ""
	currentState := state.None
	dockerVersion := "Unknown"
	hostError := ""

	url, err := h.URL()

	// PERFORMANCE: if we have the url, it's ok to assume the host is running
	// This reduces the number of calls to the drivers
	if err == nil {
		if url != "" {
			currentState = state.Running
		} else {
			currentState, err = h.Driver.GetState()
		}
	} else {
		currentState, _ = h.Driver.GetState()
	}

	if err == nil && url != "" {
		// PERFORMANCE: Reuse the url instead of asking the host again.
		// This reduces the number of calls to the drivers
		dockerHost := &mcndockerclient.RemoteDocker{
			HostURL:    url,
			AuthOption: h.AuthOptions(),
		}
		dockerVersion, err = mcndockerclient.DockerVersion(dockerHost)

		if err != nil {
			dockerVersion = "Unknown"
		} else {
			dockerVersion = fmt.Sprintf("v%s", dockerVersion)
		}
	}

	if err != nil {
		hostError = err.Error()
	}
	if hostError == drivers.ErrHostIsNotRunning.Error() {
		hostError = ""
	}

	var swarmOptions *swarm.Options
	var engineOptions *engine.Options
	if h.HostOptions != nil {
		swarmOptions = h.HostOptions.SwarmOptions
		engineOptions = h.HostOptions.EngineOptions
	}

	isMaster := false
	swarmHost := ""
	if swarmOptions != nil {
		isMaster = swarmOptions.Master
		swarmHost = swarmOptions.Host
	}

	activeHost := isActive(currentState, url)
	activeSwarm := isSwarmActive(currentState, url, isMaster, swarmHost)
	active := "-"
	if activeHost {
		active = "*"
	}
	if activeSwarm {
		active = "* (swarm)"
	}

	stateQueryChan <- commands.HostListItem{
		Name:          h.Name,
		Active:        active,
		ActiveHost:    activeHost,
		ActiveSwarm:   activeSwarm,
		DriverName:    h.Driver.DriverName(),
		State:         currentState,
		URL:           url,
		SwarmOptions:  swarmOptions,
		EngineOptions: engineOptions,
		DockerVersion: dockerVersion,
		Error:         hostError,
	}
}

func isActive(currentState state.State, hostURL string) bool {
	return currentState == state.Running && hostURL == os.Getenv("DOCKER_HOST")
}

func isSwarmActive(currentState state.State, hostURL string, isMaster bool, swarmHost string) bool {
	return isMaster && currentState == state.Running && toSwarmURL(hostURL, swarmHost) == os.Getenv("DOCKER_HOST")
}

func urlPort(urlWithPort string) string {
	parts := strings.Split(urlWithPort, ":")
	return parts[len(parts)-1]
}

func toSwarmURL(hostURL string, swarmHost string) string {
	hostPort := urlPort(hostURL)
	swarmPort := urlPort(swarmHost)
	return strings.Replace(hostURL, ":"+hostPort, ":"+swarmPort, 1)
}
