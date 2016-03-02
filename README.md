## A server for Docker Machine.

An http server that exposes the Docker Machine actions.
 
## Installation

    go get -u github.com/dgageot/docker-machine-daemon

## Build from sources

    make build
    
## Run

    ./docker-machine-daemon

## Samples

### List machines

    http --timeout 60 GET http://localhost:8080/machine

### Create machine

    http --timeout 60 --form PUT http://localhost:8080/machine/name driver=virtualbox

### Start machine

    http --timeout 60 POST http://localhost:8080/machine/name/start

### Stop machine

    http --timeout 60 POST http://localhost:8080/machine/name/stop

### Restart machine

    http --timeout 60 POST http://localhost:8080/machine/name/restart

### Remove machine

    http --timeout 60 POST http://localhost:8080/machine/name/remove

