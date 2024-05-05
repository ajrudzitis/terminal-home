package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/ajrudzitis/ssh-resume/app"
	"github.com/ajrudzitis/ssh-resume/ssh"
	log "github.com/sirupsen/logrus"
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
	a := app.ResumeApp{}
	a.Run(nil)
}

func runServer(bindIP net.IP, bindPort int64) {
	// generate a random ecdsa key pair for the server
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatalf("failed to generate server key: %v", err)
	}
	ssh.NewServer(bindIP, bindPort, privateKey, &app.ResumeApp{})
}
