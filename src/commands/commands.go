package commands

import (
	"mapset"
)

type Status uint8
type ReqType uint8

const (
	PREACCEPTED Status = iota
	ACCEPTED
	COMMITTED
)

const (
	READ ReqType = iota
	EXECUTE
	EXECUTEANDREAD
)

type Slot struct {
	ReplicaId int
	Inst      uint32
}

type Command struct {
	S     Slot
	R     ReqType
	Key   int
	Value int
	Seq   uint32
	Deps  mapset.Set
	stat  Status
}
