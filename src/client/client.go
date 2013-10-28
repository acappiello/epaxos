package main

import (
    "flag"
    "fmt"
    "strconv"

    "message"
    "util"
)

var host *string = flag.String("h", "localhost",
    "Server hostname. Default: localhost")
var port *int = flag.Int("p", 5000, "Port. Default: 5000")

func main() {
    flag.Parse()

    fmt.Println("Hello, client.")
    fmt.Println("Port is: " + strconv.Itoa(*port))

    nreq := 10
    for i := 0; i < nreq; i++ {
        fmt.Printf("%d\n", i)
        m := message.GetRequest(i)
        conn := util.NewConn(*host, *port)
        m.Marshal(conn)
        conn.Flush()
    }
}
