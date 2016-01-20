package ssh

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"regexp"
	"strings"

	"github.com/dgageot/docker-machine-daemon/daemon"
	"github.com/dgageot/docker-machine-daemon/handlers"
	"golang.org/x/crypto/ssh"
)

var (
	regexpCommandLine = regexp.MustCompile("('[^']*')|(\\S+)")

	errNoPrivateKey    = errors.New("Failed to load private key (./id_rsa). You can generate a keypair with 'ssh-keygen -t rsa -f id_rsa'")
	errParsePrivateKey = errors.New("Failed to parse private key")
)

type sshDaemon struct {
	mappings []handlers.Mapping
}

// NewDaemon create a new ssh daemon with given mappings.
func NewDaemon(mappings []handlers.Mapping) daemon.Starter {
	return &sshDaemon{
		mappings: mappings,
	}
}

// Start starts the ssh daemon.
func (d *sshDaemon) Start(port int) error {
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

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("Failed to listen on %d (%s)", port, err)
	}

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
		go d.handleChannels(chans)
	}

	return nil
}

func (d *sshDaemon) handleChannels(chans <-chan ssh.NewChannel) {
	for newChannel := range chans {
		go d.handleChannel(newChannel)
	}
}

func (d *sshDaemon) handleChannel(newChannel ssh.NewChannel) {
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
				commandLine := string(req.Payload[4:])
				parts := parseFields(commandLine)

				command := parts[0]
				args := parts[1:]

				req.Reply(true, nil)

				var output []byte = []byte("UNKNOWN")
				var err error

				for _, mapping := range d.mappings {
					if command == mapping.Url {
						output, err = handlers.ToJson(handlers.WithApi(mapping.Handler, args...))
						break
					}
				}

				if err != nil {
					fmt.Println(err)
					output = []byte("ERROR: " + err.Error())
				}

				connection.Write(output)
				connection.Close()
			}
		}
	}()
}

func parseFields(commandLine string) []string {
	fields := regexpCommandLine.FindAllString(commandLine, -1)

	for i := range fields {
		if len(fields[i]) > 2 && strings.HasPrefix(fields[i], "'") && strings.HasSuffix(fields[i], "'") {
			fields[i] = fields[i][1 : len(fields[i])-1]
		}
	}

	return fields
}
