package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/ajrudzitis/terminal-home/app"
	"github.com/ajrudzitis/terminal-home/ssh"
	"github.com/ajrudzitis/terminal-home/versioning"
	log "github.com/sirupsen/logrus"
	cryptossh "golang.org/x/crypto/ssh"
)

func main() {
	log.Infof("build verison: %s", versioning.GetBuildSha())

	bindStr := flag.String("b", "127.0.0.1", "bind address")
	portStr := flag.String("p", "2222", "port number")

	// load the server key
	keyStr := flag.String("k", "", "path to the server private key")

	flag.Parse()

	// parse the private key
	if *keyStr == "" {
		fmt.Println("server private key is required")
		os.Exit(1)
	}

	// load the private key
	pemBytes, err := os.ReadFile(*keyStr)
	if err != nil {
		fmt.Printf("failed to read private key: %v\n", err)
		os.Exit(1)
	}
	signer, err := cryptossh.ParsePrivateKey(pemBytes)
	if err != nil {
		fmt.Printf("failed to parse private key: %v\n", err)
		os.Exit(1)
	}

	// parse the bind address
	bindIP := net.ParseIP(*bindStr)
	if bindIP == nil {
		fmt.Printf("invalid bind address: %s\n", *bindStr)
		os.Exit(1)
	}

	// parse the port number
	portNum, err := strconv.ParseInt(*portStr, 10, 16)
	if err != nil {
		fmt.Printf("invalid port number: %s\n", *portStr)
	}
	log.Infof("Starting server on %s:%s\n", *bindStr, *portStr)

	// create the shell server
	runServer(bindIP, portNum, signer)
}

func runServer(bindIP net.IP, bindPort int64, signer cryptossh.Signer) {
	sshServer, err := ssh.NewServer(bindIP, bindPort, signer, &app.ResumeApp{})
	if err != nil {
		log.Fatalf("failed to create SSH server: %v", err)
	}
	sshServer.Start()
	select {}
}
