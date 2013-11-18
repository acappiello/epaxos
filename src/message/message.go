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
    T MsgType
    R ReqType
    Key int
    Rep replicainfo.ReplicaInfo
}

func AddHost(host string, port int) *Message {
    m := new(Message)
    m.T = ADDHOST
    m.Rep.Hostname = []byte(host)
    m.Rep.Port = port
    return m
}

func GetRequest(key int) *Message {
    m := new(Message)
    m.T = REQUEST
    m.R = READ
    m.Key = key
    fmt.Println(m)
    return m
}

func (m *Message) Send(wire *bufio.Writer) {
    m.Marshal(wire)
    wire.Flush()
}
