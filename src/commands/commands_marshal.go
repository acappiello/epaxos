package commands

import (
	"encoding/gob"
	"fmt"
	"io"
	"os"
)

func (t *Command) Marshal(wire io.Writer) {
	enc := gob.NewEncoder(wire)
	err := enc.Encode(t)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func (t *Command) Unmarshal(rr io.Reader) error {
	dec := gob.NewDecoder(rr)
	err := dec.Decode(t)
	return err
}
