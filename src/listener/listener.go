package listener

import (
	"bufio"
	"fmt"
	"net"
	"os"

	"message"
)

const BATCHSIZE int = 10

type Listener struct {
	messages chan message.Message
	ln       net.Listener
}

func NewListener(ln net.Listener) *Listener {
	l := new(Listener)
	l.messages = make(chan message.Message)
	l.ln = ln
	return l
}

func (l *Listener) Get() message.Message {
	return <-l.messages
}

func (l *Listener) handleConnection(conn net.Conn) {
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

func (l *Listener) Listen() {
	for {
		conn, err := l.ln.Accept()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Bad connection.")
			continue
		}
		go l.handleConnection(conn)
	}
}
