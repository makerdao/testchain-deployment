package command

import (
	"bytes"
	"fmt"
	"os/exec"
)

//Command is wrapper under exec.Cmd
type Command struct {
	exec.Cmd
	Stdout *bytes.Buffer
}

//New init wrapper
func New(cmd *exec.Cmd) *Command {
	return &Command{
		Cmd:    *cmd,
		Stdout: bytes.NewBufferString(``),
	}
}

//WithDir set workdir for command
func (c *Command) WithDir(dir string) *Command {
	c.Dir = dir
	return c
}

//WithEnvVarsMap add env vars to command from map
func (c *Command) WithEnvVarsMap(vars map[string]string) *Command {
	for env, val := range vars {
		c.Env = append(c.Env, fmt.Sprintf(`%s="%s"`, env, val))
	}
	return c
}

//Run command and use buffers for out results
func (c *Command) Run() *Error {
	bytesBuf := bytes.NewBufferString(``)
	c.Cmd.Stdout = c.Stdout
	c.Cmd.Stderr = bytesBuf
	if err := c.Cmd.Run(); err != nil {
		return NewError(err, bytesBuf.Bytes())
	}
	return nil
}
