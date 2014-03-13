package commands

import (
	"fmt"

	"datatypes"
	"stack"
)

type Graph map[datatypes.Slot] *datatypes.Vertex
type Component map[datatypes.Slot] bool

func (d *Data) BuildGraph(cmd *Command) Graph {
	g := make(Graph)
	v := new(datatypes.Vertex)
	v.Label = cmd.S
	v.Index = -1
	v.Lowlink = -1
	v.Adj = make(datatypes.Adjacencies)
	g[cmd.S] = v
	d.addGraphDeps(cmd, g)
	return g
}

func (d *Data) addGraphDeps(cmd *Command, g Graph) {
	for i, v := range(cmd.Deps) {
		S := datatypes.Slot{i, v}
		if v > 0 && !d.cmds[S].Executed {
			g[cmd.S].Adj[S] = true
			v := new(datatypes.Vertex)
			v.Label = S
			v.Index = -1
			v.Lowlink = -1
			v.Adj = make(datatypes.Adjacencies)
			g[S] = v
			d.addGraphDeps(d.cmds[S], g)
		}
	}
}

func (g *Graph) SCC() {
	index := 0
	S := stack.NewStack(100)
	for _, v := range(*g) {
		if v.Index == -1 {
			g.strongconnect(v, S, &index)
		}
	}
}

func min(x, y int) int {
	if x > y {
		return y
	}
	return x
}

func (g *Graph) strongconnect(v *datatypes.Vertex, S *stack.Stack,
	index *int) {
	v.Index = *index
	v.Lowlink = *index
	*index++
	S.Push(v)
	v.InStack = true

	for k := range(v.Adj) {
		w := (*g)[k]
		if w.Index == -1 {
			g.strongconnect(w, S, index)
			v.Lowlink = min(v.Lowlink, w.Lowlink)
		} else if (w.InStack) {
			v.Lowlink = min(v.Lowlink, w.Index)
		}
	}

	if v.Lowlink == v.Index {
		comp := make(Component)
		w := S.Pop()
		w.InStack = false
		comp[w.Label] = true
		for w.Label != v.Label {
			w = S.Pop()
			w.InStack = false
			comp[w.Label] = true
		}
		componentsToExecute <- comp
	}
}

// Don't actually need this.
func (g *Graph) buildSCC(V map[datatypes.Slot] bool) *Graph {
	c := make(Graph)
	for v := range(V) {
		w := new(datatypes.Vertex)
		w.Label = v
		w.Adj = make(datatypes.Adjacencies)
		adj := (*g)[v].Adj
		for u := range(adj) {
			if V[u] {
				w.Adj[u] = true
			}
		}
		c[v] = w
	}
	return &c
}

func (g *Graph) Print() {
	fmt.Println("===================")
	for k, v := range(*g) {
		fmt.Println(k, *v)
	}
	fmt.Println("~~~~~~~~~~~~~~~~~~~")
}
