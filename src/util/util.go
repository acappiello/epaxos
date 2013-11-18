package util

import (
	"bufio"
	"fmt"
	"net"
)

func NewConn(host string, port int) *bufio.Writer {
	conn, _ := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	return bufio.NewWriter(conn)
}
