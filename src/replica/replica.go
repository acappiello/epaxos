package main

import (
    "flag"
    "fmt"
    "net"
    "os"
    "strconv"

    "message"
    "state"
    "util"
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

    state := state.Initialize(*port, *nReplica)
    ln, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))

    if err != nil {
        fmt.Fprintf(os.Stderr, "Unable to listen on port %d\n", *port)
        return
    }

    if len(*connect) > 0 {
        fmt.Println("Registering self.")
        conn := util.NewConn(*connect, *connectPort)
        host, _ := os.Hostname()
        m := message.AddHost(host, *port)
        m.Marshal(conn)
        conn.Flush()
    }
    for i := 0; i < *nReplica-1; i++ {
        fmt.Printf("Waiting %d...\n", i)
        conn, err := ln.Accept()
        fmt.Println("Accepted.")
        if err != nil {
            fmt.Fprintln(os.Stderr, "Bad connection.")
            continue
        }
        state.NewConnection(conn)
        fmt.Printf("Complete %d...\n", i)
    }
    if len(*connect) == 0 {
        fmt.Println("Sending host info.")
        state.SendHosts()
    }
    fmt.Println("Done.")
}
