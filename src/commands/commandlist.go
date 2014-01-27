package commands

import (
	"mapset"
)

type CommandList map[Slot] *Command
type Deps map[int] mapset.Set
type NextSeq map[int] uint32

type Data struct {
	cmds CommandList
	deps Deps
	seqs NextSeq
}

func InitData(quorum int, id int) (*Data) {
	d := new(Data)
	d.cmds = make(CommandList)
	d.deps = make(Deps)
	d.seqs = make(NextSeq)
	return d
}

func max(x uint32, y uint32) uint32 {
	if x > y {
		return x
	}
	return y
}

func (d *Data) AddDepsAndSeq(cmd *Command) {
	oldDeps, exists := d.deps[cmd.Key]
	thisSeq := max(d.seqs[cmd.Key], cmd.Seq)
	d.seqs[cmd.Key] = thisSeq + 1
	cmd.Seq = thisSeq
	cmd.Deps = cmd.Deps.Union(oldDeps)

	if !exists {
		d.deps[cmd.Key] = mapset.NewSet()
	}
	d.deps[cmd.Key].Add(cmd.S)
}

func (d *Data) AddCmd(cmd *Command) {
	d.cmds[cmd.S] = cmd
}

func (d *Data) HandlePreaccept(cmd *Command) {
	union := d.deps[cmd.Key].Union(cmd.Deps)
	cmd.Deps = union
	d.deps[cmd.Key] = union.Clone()
	d.deps[cmd.Key].Add(cmd.S)
	d.cmds[cmd.S] = cmd
}

func (d *Data) HandlePreacceptOk(cmd *Command) int {
	thisCmd := d.cmds[cmd.S]
	union := d.deps[cmd.Key].Union(cmd.Deps)
	if d.deps[cmd.Key].Cardinality() < union.Cardinality() {
		thisCmd.Slow = true
	}
	d.deps[cmd.Key] = union
	thisCmd.Deps = union
	thisCmd.nOks = thisCmd.nOks + 1
	return thisCmd.nOks
}

func (d *Data) HandleAccept(cmd *Command) {
	d.deps[cmd.Key] = d.deps[cmd.Key].Union(cmd.Deps)
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
