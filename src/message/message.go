package message

import (
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
)

const (
    PUT ReqType = iota
    GET
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
    m.R = GET
    m.Key = key
    fmt.Println(m)
    return m
}
