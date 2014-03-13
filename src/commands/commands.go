package commands

import (
	"datatypes"
)

type KeyType datatypes.KeyType
type ValueType datatypes.ValueType

type ReqType uint8

const (
	READ ReqType = iota
	EXECUTE
	EXECUTEANDREAD
)

type Command struct {
	S         datatypes.Slot
	R         ReqType
	Key       KeyType
	Value     ValueType
	Seq       uint32
	Deps      []uint32
	NOks      int
	Slow      bool
	Accepted  bool
	Committed bool
	Executed  bool
	ClientId  int64
}
