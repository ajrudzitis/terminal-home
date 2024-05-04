package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	tcell "github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

func main() {
	// if no arguments are given, run locally
	if len(os.Args) == 1 {
		runLocal()
		return
	}

	// if some arguments are given, run the server

	port := "2222"
	// address read from the command line
	bindAddr := os.Args[1]
	// optional port read from the command line
	if len(os.Args) > 2 {
		port = os.Args[2]
	}

	bindIP := net.ParseIP(bindAddr)
	if bindIP == nil {
		fmt.Printf("invalid bind address: %s\n", bindAddr)
		os.Exit(1)
	}

	portNum, err := strconv.ParseInt(port, 10, 16)
	if err != nil {
		fmt.Printf("invalid port number: %s\n", port)
	}
	log.Infof("Starting server on %s:%s\n", bindAddr, port)

	// create the shell server
	runServer(bindIP, portNum)
}

func runLocal() {
	screen, err := tcell.NewScreen()
	if err != nil {
		fmt.Printf("error: %v", err)
		os.Exit(1)
	}
	runApp(screen)
}

func runServer(bindIP net.IP, bindPort int64) {
	// generate a random ecdsa key pair for the server
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatalf("failed to generate server key: %v", err)
	}
	signer, err := ssh.NewSignerFromKey(privateKey)
	if err != nil {
		log.Fatalf("failed to create signer: %v", err)
	}

	// create a server config
	config := &ssh.ServerConfig{
		NoClientAuth: true,
	}
	config.AddHostKey(signer)

	// create a listener on a random port
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{IP: bindIP, Port: int(bindPort)})
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// log the listender
	log.Infof("SSH connection listening on %s", listener.Addr().String())

	for {

		// set a deadline for the listener to accept a connection
		listener.SetDeadline(time.Now().Add(5 * time.Second))
		// accept a connection
		conn, err := listener.Accept()
		if err != nil {
			// if the error is due to a timeout, continue to the next iteration
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				continue
			}
			return
		}

		log.Infof("accepted connection from %s", conn.RemoteAddr().String())

		go handleConnection(conn, config)
	}
}

func handleConnection(conn net.Conn, config *ssh.ServerConfig) {
	// handle the connection
	server, chans, reqs, err := ssh.NewServerConn(conn, config)
	if err != nil {
		fmt.Printf("failed to handshake: %v\n", err)
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
			fmt.Printf("could not accept channel: %v\n", err)
			return
		}

		pty := &sshtty{
			channel: channel,
		}
		screen, err := tcell.NewTerminfoScreenFromTty(pty)
		if err != nil {
			log.Errorf("failed to create terminfo screen: %v", err)
		}

		// handle requests on the channel
		go func() {
			// we're okay with this being a shell or pty request
			for req := range requests {
				ok := false
				switch req.Type {
				case "shell":
					ok = true
				case "pty-req":
					ok = true
					ptyReq := &ptyRequestMsg{}
					err := ssh.Unmarshal(req.Payload, ptyReq)
					if err != nil {
						log.Errorf("failed to unmarshal pty request: %v", err)
					}
					pty.UpdateWindow(tcell.WindowSize{Width: int(ptyReq.Columns), Height: int(ptyReq.Rows), PixelWidth: int(ptyReq.Width), PixelHeight: int(ptyReq.Height)})
					log.Infof("pty request: %v", ptyReq)
				}
				req.Reply(ok, nil)
			}
		}()

		log.Info("starting application session")

		runApp(screen)

		// close the channel
		channel.Close()
	}
	server.Close()
	log.Infof("completed session from %s", conn.RemoteAddr().String())
}

func runApp(screen tcell.Screen) {
	app := tview.NewApplication()
	app.SetScreen(screen)
	form := tview.NewForm().
		AddInputField("Enter Card Number", "", 16, nil, nil).
		AddButton("Submit", nil).
		AddButton("Quit", func() {
			app.Stop()
		})
	form.SetBorder(true).SetTitle("Enter Payment Details").SetTitleAlign(tview.AlignLeft)
	if err := app.SetRoot(form, true).Run(); err != nil {
		log.Errorf("failed to run app: %v", err)
	}
}

type ptyRequestMsg struct {
	Term     string
	Columns  uint32
	Rows     uint32
	Width    uint32
	Height   uint32
	Modelist string
}

type sshtty struct {
	channel    ssh.Channel
	windowSize tcell.WindowSize
	windowCb   func()
}

// Close implements tcell.Tty.
func (s *sshtty) Close() error {
	return s.channel.Close()
}

// Read implements tcell.Tty.
func (s *sshtty) Read(p []byte) (n int, err error) {
	log.Infof("pty read %d bytes", len(p))
	return s.channel.Read(p)
}

// Write implements tcell.Tty.
func (s *sshtty) Write(p []byte) (n int, err error) {
	log.Infof("pty write %d bytes", len(p))
	return s.channel.Write(p)
}

func (s *sshtty) UpdateWindow(size tcell.WindowSize) {
	log.Infof("updating window size: %v", size)
	s.windowSize = size
	if s.windowCb != nil {
		s.windowCb()
	}
}

// Drain implements tcell.Tty.
func (s *sshtty) Drain() error {
	return nil
}

// NotifyResize implements tcell.Tty.
func (s *sshtty) NotifyResize(cb func()) {
	s.windowCb = cb
}

// Start implements tcell.Tty.
func (s *sshtty) Start() error {
	return nil
}

// Stop implements tcell.Tty.
func (s *sshtty) Stop() error {
	return nil
}

// WindowSize implements tcell.Tty.
func (s *sshtty) WindowSize() (tcell.WindowSize, error) {
	log.Infof("checking window size: %v", s.windowSize)
	return s.windowSize, nil
}

var _ tcell.Tty = &sshtty{}
