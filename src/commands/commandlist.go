package commands

import (
	"mapset"
)

type CommandList map[Slot] Command
type Deps map[int] mapset.Set
type NextSeq map[int] uint32

type Data struct {
	cmds CommandList
	deps Deps
	seqs NextSeq
}

func InitData() (*Data) {
	d := new(Data)
	d.cmds = make(CommandList)
	d.deps = make(Deps)
	d.seqs = make(NextSeq)
	return d
}

func (d *Data) AddDepsAndSeq(cmd *Command) {
	oldDeps, exists := d.deps[cmd.Key]
	thisSeq := d.seqs[cmd.Key]
	d.seqs[cmd.Key] = thisSeq + 1
	cmd.Seq = thisSeq
	cmd.Deps = cmd.Deps.Union(oldDeps)

	if !exists {
		d.deps[cmd.Key] = mapset.NewSet()
	}
	d.deps[cmd.Key].Add(cmd.S)
}
