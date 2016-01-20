export GO15VENDOREXPERIMENT = 1

.DEFAULT_GOAL := build

build: docker-machine-daemon

docker-machine-daemon: main.go daemon/*.go daemon/ssh/*.go daemon/http/*.go handlers/*.go
	go build .

deps:
	godep save

clean:
	rm docker-machine-daemon

