package message

import (
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
    Rep replicainfo.ReplicaInfo
}

func AddHost(host string, port int) *Message {
    m := new(Message)
    m.T = ADDHOST
    m.Rep.Hostname = []byte(host)
    m.Rep.Port = port
    return m
}
