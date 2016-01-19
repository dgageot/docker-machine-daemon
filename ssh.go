package main

import (
	"fmt"
	"log"

	"io/ioutil"
	"net"

	"errors"

	"golang.org/x/crypto/ssh"
)

var (
	errNoPrivateKey    = errors.New("Failed to load private key (./id_rsa). You can generate a keypair with 'ssh-keygen -t rsa -f id_rsa'")
	errParsePrivateKey = errors.New("Failed to parse private key")
)

func startSshDaemon(port int) error {
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

	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
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
