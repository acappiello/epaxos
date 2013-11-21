// Package message contains data structures to keep track of tasks and messages
// that move between replicas as well as clients.
package message

import (
	"bufio"
	"fmt"

	"replicainfo"
)

type MsgType uint8
type ReqType uint8

const (
	REQUEST MsgType = iota
	CONNECT
	HOSTLIST
	ADDHOST
	PREACCEPT
	PREACCEPTOK
	ACCEPT
	ACCEPTOK
	COMMIT
)

const (
	READ ReqType = iota
	EXECUTE
	EXECUTEANDREAD
)

type Message struct {
	T     MsgType
	R     ReqType
	Rep   replicainfo.ReplicaInfo
	// All tasks in the same message must be from the same source and be
	// of the same type.
	Tasks []Task
}

// AddHost creates a message for a replica that needs to join the group.
func AddHost(host string, port int) *Message {
	m := new(Message)
	m.T = ADDHOST
	m.Rep.Hostname = []byte(host)
	m.Rep.Port = port
	return m
}

// ReadRequest creates a message for a READ request from a client.
func ReadRequest(key int) *Message {
	m := new(Message)
	m.T = REQUEST
	m.R = READ
	m.Tasks = make([]Task, 1)
	m.Tasks[0].Key = key
	fmt.Println(m)
	return m
}

// Send is a shortcut to send a message and flush the buffer.
func (m *Message) Send(wire *bufio.Writer) {
	m.Marshal(wire)
	wire.Flush()
}

type Task struct {
	Key    int
	HostId int
}
