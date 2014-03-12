package commands

type KeyType int
type ValueType int

type ReqType uint8

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
	S         Slot
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
