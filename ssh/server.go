package ssh

import (
	"fmt"
	"net"
	"time"

	pkgPty "github.com/ajrudzitis/ssh-resume/pty"
	"github.com/gdamore/tcell/v2"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

type sshServer struct {
	listener        *net.TCPListener
	sshServerConfig *ssh.ServerConfig
	app             SSHApplication
}

// NewServer creates a new SSH server
func NewServer(bindIP net.IP, bindPort int64, privateKey interface{}, app SSHApplication) (*sshServer, error) {
	// create a signer from the private key
	signer, err := ssh.NewSignerFromKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create signer from private key: %w", err)
	}

	// create a server config
	config := &ssh.ServerConfig{
		NoClientAuth: true,
	}
	config.AddHostKey(signer)

	// create a listener on the specified port
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{IP: bindIP, Port: int(bindPort)})
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s:%d: %w", bindIP, bindPort, err)
	}

	return &sshServer{listener: listener, sshServerConfig: config, app: app}, nil
}

// Start starts accepting connections. It returns immediately.
func (s *sshServer) Start() {
	go s.acceptConnections()
}

// acceptConnections listens for incoming connections and handles them
// this method blocks until the listener is closed
func (s *sshServer) acceptConnections() {
	for {
		// set a deadline for the listener to accept a connection
		// this gives us a chance to abort
		// TODO: make the deadline configurable
		s.listener.SetDeadline(time.Now().Add(5 * time.Second))
		// accept a connection
		conn, err := s.listener.Accept()
		if err != nil {
			// if the error is due to a timeout, continue to the next iteration
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				continue
			}
			return
		}

		log.Infof("ssh: accepted connection from %s", conn.RemoteAddr().String())

		// handle the connection in a goroutine
		go s.handleConnection(conn)
	}
}

// handleConnection handles an incoming connection. It blocks until the connection is closed
func (s *sshServer) handleConnection(conn net.Conn) {
	// spin up a new server to handle the connection
	server, chans, reqs, err := ssh.NewServerConn(conn, s.sshServerConfig)
	if err != nil {
		log.Warnf("ssh: failed to handshake: %v", err)
		return
	}

	// we can disregard basic requests
	go ssh.DiscardRequests(reqs)

	// handle channel requests
	for newChannel := range chans {

		// only accept session channels
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}

		// accept the channel
		channel, requests, err := newChannel.Accept()
		if err != nil {
			log.Warnf("ssh: could not accept channel: %v", err)
			return
		}

		pty := pkgPty.NewPty(channel)

		// handle requests on the channel
		go func() {
			// we're okay with this being a shell or pty request
			for req := range requests {
				ok := false
				switch req.Type {
				case "shell":
					ok = true
				case "pty-req":
					// we need to unmarshal the request to get the window size
					ok = true
					ptyReq := &pkgPty.PtyRequestMsg{}

					err := ssh.Unmarshal(req.Payload, ptyReq)
					if err != nil {
						log.Warnf("ssh: failed to unmarshal pty request: %v", err)
					}
					pty.UpdateWindow(tcell.WindowSize{Width: int(ptyReq.Columns), Height: int(ptyReq.Rows), PixelWidth: int(ptyReq.Width), PixelHeight: int(ptyReq.Height)})
				}
				req.Reply(ok, nil)
			}
		}()

		log.Infof("ssh: starting application session for %s", conn.RemoteAddr().String())

		s.app.Run(pty)

		// close the channel
		channel.Close()
	}
	server.Close()
	log.Infof("ssh: completed session from %s", conn.RemoteAddr().String())
}
