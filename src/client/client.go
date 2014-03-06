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

func send(nreq int, w *bufio.Writer) {
	for i := 0; i < nreq; i++ {
		m := message.ReadRequest(i % 100)
		fmt.Println("SEND: ", i, m)
		m.Send(w)
	}
}

func main() {
	flag.Parse()

	fmt.Println("Hello, client.")
	fmt.Println("Port is: " + strconv.Itoa(*port))

	nreq := 10000
	conn, _ := net.Dial("tcp", fmt.Sprintf("%s:%d", *host, *port))
	buf := bufio.NewWriter(conn)
	reader := bufio.NewReader(conn)

	go send(nreq, buf)

	rep := &message.Message{}
	for i := 0; i < nreq; i++ {
		rep.Unmarshal(reader)
		fmt.Println("DONE: ", i, rep)
	}

	conn.Close()
}
