package main

import (
	"encoding/json"
	"fmt"
	"log"

	"io/ioutil"
	"net"

	"github.com/dgageot/docker-machine-daemon/ls"
	"github.com/docker/machine/commands/mcndirs"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/persist"

	"errors"

	"net/http"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/ssh"
)

const (
	sshPort  = 2200
	httpPort = 8080
)

var (
	errNoPrivateKey    = errors.New("Failed to load private key (./id_rsa). You can generate a keypair with 'ssh-keygen -t rsa -f id_rsa'")
	errParsePrivateKey = errors.New("Failed to parse private key")
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

func startHttpServer() error {
	log.Printf("Listening on %d...\n", httpPort)
	log.Printf(" - List the Docker Machines with: http GET http://localhost:%d/machine/ls\n", httpPort)

	r := mux.NewRouter()
	r.HandleFunc("/machine/ls", toHandlerFunc(runLs))

	http.ListenAndServe(fmt.Sprintf(":%d", httpPort), r)

	return nil
}

func toHandlerFunc(handler func(api libmachine.API) (interface{}, error)) http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		output, err := toJson(withApi(handler))
		if err != nil {
			response.WriteHeader(500)
			return
		}

		response.Write(output)
	}
}

func startSshDaemon() error {
	privateBytes, err := ioutil.ReadFile("id_rsa")
	if err != nil {
		return errNoPrivateKey
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		return errParsePrivateKey
	}

	config := &ssh.ServerConfig{
		NoClientAuth: true,
	}

	config.AddHostKey(private)

	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", sshPort))
	if err != nil {
		return fmt.Errorf("Failed to listen on %d (%s)", sshPort, err)
	}

	log.Printf("Listening on %d...\n", sshPort)
	log.Printf(" - List the Docker Machines with: ssh localhost -p %d -s machine/ls\n", sshPort)
	for {
		tcpConn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept incoming connection (%s)", err)
			continue
		}

		sshConn, chans, reqs, err := ssh.NewServerConn(tcpConn, config)
		if err != nil {
			log.Printf("Failed to handshake (%s)", err)
			continue
		}

		log.Printf("New SSH connection from %s (%s)", sshConn.RemoteAddr(), sshConn.ClientVersion())
		go ssh.DiscardRequests(reqs)
		go handleChannels(chans)
	}

	return nil
}

func handleChannels(chans <-chan ssh.NewChannel) {
	for newChannel := range chans {
		go handleChannel(newChannel)
	}
}

func handleChannel(newChannel ssh.NewChannel) {
	if t := newChannel.ChannelType(); t != "session" {
		newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
		return
	}

	connection, requests, err := newChannel.Accept()
	if err != nil {
		log.Printf("Could not accept channel (%s)", err)
		return
	}

	go func() {
		for req := range requests {
			switch req.Type {
			case "subsystem":
				command := string(req.Payload[4:])
				req.Reply(true, nil)

				var output []byte
				var err error
				if command == "machine/ls" {
					output, err = toJson(withApi(runLs))
				} else {
					fmt.Println(command)
					output = []byte("UNKNOWN")
				}

				if err != nil {
					fmt.Println(err)
					output = []byte("ERROR: " + err.Error())
				}

				connection.Write(output)
				connection.Close()
				log.Printf("Session closed")
			}
		}
	}()
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
