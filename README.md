## A server for Docker Machine.

Currently I play with two options:

 + An http server
 + An ssh server using ssh subsystems
 
## Installation

    GO15VENDOREXPERIMENT=1 go get github.com/dgageot/docker-machine-daemon

## Build from sources

    make build
    
## Run

    ./docker-machine-daemon