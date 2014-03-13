package datatypes

type KeyType int
type ValueType int

type Slot struct {
	ReplicaId int
	Inst      uint32
}

type Adjacencies map[Slot] bool

type Vertex struct {
	Label   Slot
	Index   int
	Lowlink int
	InStack bool
	Adj     Adjacencies
}
