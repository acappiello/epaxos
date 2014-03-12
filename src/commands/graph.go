package commands

type Adjacencies map[Slot] bool
type Graph map[Slot] Adjacencies

func (d *Data) BuildGraph(cmd *Command) Graph {
	//fmt.Println(cmd.Deps)
	g := make(Graph)
	g[cmd.S] = make(Adjacencies)
	d.addGraphDeps(cmd, g)
	return g
}

func (d *Data) addGraphDeps(cmd *Command, g Graph) {
	for i, v := range(cmd.Deps) {
		S := Slot{i, v}
		if v > 0 && !d.cmds[S].Executed {
			g[cmd.S][S] = true
			g[S] = make(Adjacencies)
			d.addGraphDeps(d.cmds[S], g)
		}
	}
}
