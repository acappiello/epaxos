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

type Commands []*Command

func (A Commands) Len() int {
	return len(A)
}

func (A Commands) Swap(i, j int) {
	tmp := A[i]
	A[i] = A[j]
	A[j] = tmp
}

func (A Commands) Less(i, j int) bool {
	return A[i].Seq < A[j].Seq
}
