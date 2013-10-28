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

    nreq := 10
    conn, _ := net.Dial("tcp", fmt.Sprintf("%s:%d", *host, *port))
    buf := bufio.NewWriter(conn)
    for i := 0; i < nreq; i++ {
        m := message.GetRequest(i)
        m.Marshal(buf)
        buf.Flush()
    }
    conn.Close()
}
