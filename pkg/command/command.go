package command

import (
	"bytes"
	"os/exec"
)

type Command struct {
	exec.Cmd
	Stdout *bytes.Buffer
}

func New(cmd *exec.Cmd) *Command {
	return &Command{
		Cmd:    *cmd,
		Stdout: bytes.NewBufferString(``),
	}
}

func (c *Command) WithDir(dir string) *Command {
	c.Dir = dir
	return c
}

func (c *Command) Run() *Error {
	bytesBuf := bytes.NewBufferString(``)
	c.Cmd.Stdout = c.Stdout
	c.Cmd.Stderr = bytesBuf
	if err := c.Cmd.Run(); err != nil {
		return NewError(err, bytesBuf.Bytes())
	}
	return nil
}
