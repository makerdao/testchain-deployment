package dapp

import (
	"os/exec"

	"github.com/makerdao/testchain-deployment/pkg/command"
)

type Client struct {
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) UpdateCmd(dir string) *command.Error {
	return command.New(
		exec.Command("dapp", "update"),
	).
		WithDir(dir).
		Run()
}
