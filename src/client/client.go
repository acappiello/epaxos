// Package main is the entry point for a client.
// The workload isn't very interesting right now.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"strconv"

	"message"
)

var host *string = flag.String("h", "localhost",
	"Server hostname. Default: localhost")
var port *int = flag.Int("p", 5000, "Port. Default: 5000")

func main() {
	flag.Parse()

	fmt.Println("Hello, client.")
	fmt.Println("Port is: " + strconv.Itoa(*port))

	nreq := 100
	conn, _ := net.Dial("tcp", fmt.Sprintf("%s:%d", *host, *port))
	buf := bufio.NewWriter(conn)
	for i := 0; i < nreq; i++ {
		m := message.ReadRequest(i % 10)
		m.Send(buf)
	}
	rep := &message.Message{}
	reader := bufio.NewReader(conn)
	for i := 0; i < nreq; i++ {
		rep.Unmarshal(reader)
		fmt.Println("DONE: ", i, rep)
	}
	conn.Close()
}
