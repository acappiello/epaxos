package main

import (
    "flag"
    "fmt"
    "strconv"
)

var port *int = flag.Int("p", 5000, "Port. Default: 5000")

func main() {
    flag.Parse()

    fmt.Println("Hello, client.")
    fmt.Println("Port is: " + strconv.Itoa(*port))
}
