export GO15VENDOREXPERIMENT = 1

.DEFAULT_GOAL := build

build: docker-machine-daemon

docker-machine-daemon: main.go ls/ls.go http.go ssh.go handlers/mapping.go
	go build .

deps:
	godep save

clean:
	rm docker-machine-daemon

