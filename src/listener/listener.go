package listener

import (
	"bufio"
	"fmt"
	"net"
	"os"

	"message"
)

type Listener struct {
	messages chan message.Message
	ln       net.Listener
}

// NewListener sets the initial state of to receive input.
func NewListener(ln net.Listener) *Listener {
	l := new(Listener)
	l.messages = make(chan message.Message)
	l.ln = ln
	return l
}

// Get gives the next message, blocking if needed.
func (l *Listener) Get() message.Message {
	return <-l.messages
}

// HandleConnection serves a single connection until error or disconnect.
func (l *Listener) HandleConnection(conn net.Conn) {
	m := &message.Message{}
	buffered := bufio.NewReader(conn)
	for {
		err := m.Unmarshal(buffered)
		if err != nil {
			fmt.Println("Read error: ", err)
			return
		}
		l.messages <- *m
	}
}

// Listen accepts new connections forever and starts a new goroutine for each.
func (l *Listener) Listen() {
	for {
		conn, err := l.ln.Accept()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Bad connection.")
			continue
		}
		go l.HandleConnection(conn)
	}
}
