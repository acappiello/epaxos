package commands

type CommandList map[Slot]*Command
type Deps map[int][]uint32

type Data struct {
	cmds     CommandList
	deps     Deps
	nreplica int
	id       int
}

func InitData(nreplica int, id int) *Data {
	d := new(Data)
	d.cmds = make(CommandList)
	d.deps = make(Deps)
	d.nreplica = nreplica
	d.id = id
	return d
}

func max(x, y uint32) uint32 {
	if x > y {
		return x
	}
	return y
}

func maxArr(A []uint32) uint32 {
	var x uint32 = 0
	for _, v := range A {
		if v > x {
			x = v
		}
	}
	return x
}

func (d *Data) AddDepsAndSeq(cmd *Command) {
	deps, exists := d.deps[cmd.Key]
	if !exists {
		deps = make([]uint32, d.nreplica)
		d.deps[cmd.Key] = deps
	}

	thisSeq := maxArr(deps) + 1
	deps[d.id] = thisSeq
	cmd.Deps = make([]uint32, d.nreplica)
	copy(cmd.Deps, deps)
}

func (d *Data) AddCmd(cmd *Command) {
	d.cmds[cmd.S] = cmd
}

func updateDeps(D, S []uint32) uint32 {
	var diff uint32 = 0
	for i, v := range S {
		if v > D[i] {
			D[i] = v
			diff++
		}
	}
	return diff
}

func updateDepsBi(D, S []uint32) uint32 {
	var diff uint32 = 0
	for i, v := range S {
		if v > D[i] {
			D[i] = v
			diff++
		} else if v < D[i] {
			S[i] = D[i]
			diff++
		}
	}
	return diff
}

func (d *Data) HandlePreaccept(cmd *Command) {
	deps, exists := d.deps[cmd.Key]
	if !exists {
		deps = make([]uint32, d.nreplica)
		d.deps[cmd.Key] = deps
	}
	seq := maxArr(deps)
	seq = max(seq, cmd.Seq)
	cmd.Seq = seq
	deps[cmd.S.ReplicaId] = seq
	updateDepsBi(deps, cmd.Deps)
	d.cmds[cmd.S] = cmd
}

func (d *Data) HandlePreacceptOk(cmd *Command) int {
	thisCmd := d.cmds[cmd.S]
	if updateDepsBi(thisCmd.Deps, cmd.Deps) > 0 {
		thisCmd.Slow = true
	}
	updateDeps(d.deps[cmd.Key], cmd.Deps)
	thisCmd.nOks = thisCmd.nOks + 1
	return thisCmd.nOks
}

func (d *Data) HandleAccept(cmd *Command) {
	updateDeps(d.deps[cmd.Key], cmd.Deps)
	d.cmds[cmd.S] = cmd
}

func (d *Data) HandleAcceptOk(cmd *Command) int {
	thisCmd := d.cmds[cmd.S]
	thisCmd.nOks = thisCmd.nOks + 1
	return thisCmd.nOks
}

func (d *Data) HandleCommit(cmd *Command) {
	thisCmd := d.cmds[cmd.S]
	thisCmd.Committed = true
}
