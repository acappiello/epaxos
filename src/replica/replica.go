// Package main is the entry point for a replica.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"

	"message"
	"state"
)

var port *int = flag.Int("p", 5000, "Port. Default: 5000")
var nReplica *int = flag.Int("n", 3, "Number of replicas. Default: 3")
var connect *string = flag.String("h", "",
	"Initial connection. If not specified, wait for others. Default: \"\"")
var connectPort *int = flag.Int("hp", 5000,
	"Initial connection port. Default: 5000")

func main() {
	flag.Parse()

	fmt.Println("Hello, replica.")
	fmt.Println("Port is: " + strconv.Itoa(*port))

	state, err := state.Initialize(*port, *nReplica)
	//ln, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to listen on port %d\n", *port)
		return
	}

	if len(*connect) == 0 {
		fmt.Println("Waiting for peers.")
		state.WaitForPeers()
		fmt.Println("Sending host info.")
		state.SendHosts()
	} else {
		fmt.Println("Registering self.")
		conn, _ := net.Dial("tcp", fmt.Sprintf("%s:%d", *connect, *connectPort))
		buf := bufio.NewWriter(conn)
		host, _ := os.Hostname()
		m := message.AddHost(host, *port)
		m.Send(buf)
		fmt.Println("Getting Peers.")
		state.GetPeers(conn)
	}
	fmt.Println("Done.")

	state.Run()
}
