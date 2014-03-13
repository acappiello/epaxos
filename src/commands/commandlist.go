package commands

import (
	"datatypes"
)

type CommandList map[datatypes.Slot]*Command
type Deps map[KeyType][]uint32

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

func (d *Data) maxSeq(A []uint32) uint32 {
	var max uint32 = 0
	for i, v := range(A) {
		if v > 0 {
			S := datatypes.Slot{i, v}
			seq := d.cmds[S].Seq
			if seq > max {
				max = seq
			}
		}
	}
	return max
}

func (d *Data) AddDepsAndSeq(cmd *Command) {
	deps, exists := d.deps[cmd.Key]
	if !exists {
		deps = make([]uint32, d.nreplica)
		d.deps[cmd.Key] = deps
	}

	cmd.Seq = d.maxSeq(deps) + 1
	cmd.Deps = make([]uint32, d.nreplica)
	copy(cmd.Deps, deps)
	deps[d.id] = cmd.S.Inst
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
	seq := d.maxSeq(deps) + 1
	seq = max(seq, cmd.Seq)
	cmd.Seq = seq
	updateDepsBi(deps, cmd.Deps)
	d.cmds[cmd.S] = cmd
}

func (d *Data) HandlePreacceptOk(cmd *Command) *Command {
	thisCmd := d.cmds[cmd.S]
	if updateDeps(thisCmd.Deps, cmd.Deps) > 0 {
		thisCmd.Slow = true
	}
	if thisCmd.Seq < cmd.Seq {
		thisCmd.Seq = cmd.Seq
		thisCmd.Slow = true
	}
	updateDeps(d.deps[cmd.Key], thisCmd.Deps)
	thisCmd.NOks = thisCmd.NOks + 1
	return thisCmd
}

func (d *Data) HandleAccept(cmd *Command) {
	updateDeps(d.deps[cmd.Key], cmd.Deps)
	d.cmds[cmd.S] = cmd
}

func (d *Data) HandleAcceptOk(cmd *Command) int {
	thisCmd := d.cmds[cmd.S]
	thisCmd.NOks = thisCmd.NOks + 1
	return thisCmd.NOks
}

func (d *Data) HandleCommit(cmd *Command) {
	thisCmd := d.cmds[cmd.S]
	thisCmd.Committed = true
}
