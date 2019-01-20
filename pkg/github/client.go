package github

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/makerdao/testchain-deployment/pkg/command"
	"github.com/sirupsen/logrus"
)

const (
	cloneTmplt = "git@github.com:%s/%s.git"
)

//Client of github.com
type Client struct {
	cfg     Config
	baseDir string
}

//NewClient init client
func NewClient(cfg Config, baseDir string) *Client {
	return &Client{
		cfg:     cfg,
		baseDir: baseDir,
	}
}

//GetRepoName return name of repo
func (c *Client) GetRepoPath() string {
	return filepath.Join(c.baseDir, c.cfg.RepoName)
}

func (c *Client) CleanIfExists(log *logrus.Entry) error {
	if _, err := os.Stat(c.GetRepoPath()); os.IsNotExist(err) {
		return nil
	}
	if err := os.RemoveAll(c.GetRepoPath()); err != nil {
		return err
	}
	return nil
}

func (c *Client) CloneCmd(log *logrus.Entry) *command.Error {
	return command.New(
		exec.Command("git", "clone", fmt.Sprintf(cloneTmplt, c.cfg.RepoOwner, c.cfg.RepoName)),
	).
		WithDir(c.baseDir).
		Run()
}

func (c *Client) UpdateSubmodulesCmd(log *logrus.Entry) *command.Error {
	return command.New(
		exec.Command("git", "submodule", "update", "--init"),
	).
		WithDir(c.GetRepoPath()).
		Run()
}

func (c *Client) CheckoutCmd(log *logrus.Entry, id *string) *command.Error {
	if id == nil {
		id = new(string)
		*id = "tags/" + c.cfg.TagName
	}
	return command.New(
		exec.Command("git", "checkout", *id),
	).
		WithDir(c.GetRepoPath()).
		Run()
}

func (c *Client) LastHashCommitCmd(log *logrus.Entry) (string, *command.Error) {
	cmd := command.New(
		exec.Command("git", "rev-parse", "HEAD"),
	)
	if err := cmd.WithDir(c.GetRepoPath()).Run(); err != nil {
		return "", err
	}
	return cmd.Stdout.String(), nil
}
