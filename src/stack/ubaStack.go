package stack

import (
	"fmt"
	"log"

	"datatypes"
)

type ElemType *datatypes.Vertex

type Stack struct {
	elems []ElemType
	next  uint32
	limit uint32
}

func NewStack(initialLimit uint32) (*Stack) {
	S := new(Stack)
	S.elems = make([]ElemType, initialLimit)
	S.next = 0
	S.limit = initialLimit
	return S
}

func (S *Stack) Push(e ElemType) {
	if S.next == S.limit {
		S.limit *= 2
		newStack := make([]ElemType, S.limit)
		copy(newStack, S.elems)
		S.elems = newStack
	}
	S.elems[S.next] = e
	S.next++
}

func (S *Stack) Pop() ElemType {
	if S.next == 0 {
		log.Panicln("Pop from empty stack.")
	}
	S.next--
	e := S.elems[S.next]
	if S.next == S.limit / 4 && S.limit / 2 > 0 {
		S.limit /= 2
		newStack := make([]ElemType, S.limit)
		copy(newStack, S.elems)
		S.elems = newStack
	}
	return e
}

func (S *Stack) IsEmpty() bool {
	return S.next != 0
}

func (S *Stack) Print() {
	fmt.Println("===================")
	for i := int32(S.next-1); i >= 0; i-- {
		fmt.Println("-------------------", i)
		fmt.Println(S.elems[i])
	}
	fmt.Println("~~~~~~~~~~~~~~~~~~~")
}
