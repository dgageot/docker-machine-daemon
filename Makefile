BIN := docker-machine-daemon

export GO15VENDOREXPERIMENT = 1

.DEFAULT_GOAL := build

run: build
	./$(BIN)

build: $(BIN)

docker-machine-daemon: main.go daemon/*.go daemon/ssh/*.go daemon/http/*.go handlers/*.go
	go build .

deps:
	godep save

clean:
	rm $(BIN)

