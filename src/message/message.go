package message

type MsgType uint8

const (
    PUT MsgType = iota
    GET
    CONNECT
    HOSTLIST
)

type Message struct {
    T MsgType
}
