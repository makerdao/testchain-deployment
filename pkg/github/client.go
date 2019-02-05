package github

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/makerdao/testchain-deployment/pkg/command"
	"github.com/sirupsen/logrus"
)

const (
	cloneTmplt = "https://github.com/%s/%s.git"
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

//GetRepoName return name of repo for work
func (c *Client) GetRepoPath() string {
	return filepath.Join(c.baseDir, c.cfg.RepoName)
}

//GetLoadingPath return name of repo for loading
func (c *Client) GetLoadingPath() string {
	return filepath.Join(c.baseDir, c.cfg.RepoName+"-loading")
}

func (c *Client) CleanRepoIfExists(log *logrus.Entry) error {
	if _, err := os.Stat(c.GetRepoPath()); os.IsNotExist(err) {
		return nil
	}
	if err := os.RemoveAll(c.GetRepoPath()); err != nil {
		return err
	}
	return nil
}

//CleanLoadingRepoIfExists clean path for loading repo
func (c *Client) CleanLoadingRepoIfExists(log *logrus.Entry) error {
	if _, err := os.Stat(c.GetLoadingPath()); os.IsNotExist(err) {
		return nil
	}
	if err := os.RemoveAll(c.GetLoadingPath()); err != nil {
		return err
	}
	return nil
}

//CloneCmd is git clone command
func (c *Client) CloneCmd(log *logrus.Entry) *command.Error {
	return command.New(
		exec.Command(
			"git",
			"clone",
			fmt.Sprintf(cloneTmplt, c.cfg.RepoOwner, c.cfg.RepoName),
			c.GetLoadingPath(),
		),
	).
		WithDir(c.baseDir).
		Run()
}

//UpdateSubmodulesCmd run update submodules
func (c *Client) UpdateSubmodulesCmd(log *logrus.Entry) *command.Error {
	return command.New(
		exec.Command("git", "submodule", "update", "--init", "--recursive"),
	).
		WithDir(c.GetLoadingPath()).
		Run()
}

//CheckoutCmd checkout to id(for example tag or branch), if empty checkout to tag from settings
func (c *Client) CheckoutCmd(log *logrus.Entry, id *string) *command.Error {
	if id == nil {
		id = new(string)
		*id = "tags/" + c.cfg.TagName
	}
	return command.New(
		exec.Command("git", "checkout", *id),
	).
		WithDir(c.GetLoadingPath()).
		Run()
}

//LastHashCommitCmd return string with hash commit for tag
func (c *Client) LastHashCommitCmd(log *logrus.Entry) (string, *command.Error) {
	cmd := command.New(
		exec.Command("git", "rev-parse", "HEAD"),
	)
	if err := cmd.WithDir(c.GetRepoPath()).Run(); err != nil {
		return "", err
	}
	return strings.TrimSuffix(cmd.Stdout.String(), "\n"), nil
}
