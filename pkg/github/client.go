package github

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/makerdao/testchain-deployment/pkg/command"
)

const (
	cloneTmplt = "https://github.com/%s/%s.git"
)

//Client of github.com
type Client struct {
	cfg              Config
	baseDir          string
	commitListRegexp *regexp.Regexp
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
		*id = c.cfg.DefaultCheckoutTarget
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

type Commit struct {
	Ref    string `json:"ref"`
	Commit string `json:"commit"`
	Author string `json:"author"`
	Date   string `json:"date"`
	Text   string `json:"text"`
}

const parseCommitRegexp = `(?m)^(.*)\|(.*)\|(.*)\|(.*)\|(.*)$`

func (c *Client) GetCommitList(log *logrus.Entry) ([]Commit, *command.Error, error) {
	if c.commitListRegexp == nil {
		var err error
		c.commitListRegexp, err = regexp.Compile(parseCommitRegexp)
		if err != nil {
			return nil, nil, err
		}
	}

	cmd := command.New(
		exec.Command("git", "log", "--all", "--pretty=format:%D|%H|%an <%ae>|%aI|%s"),
	)
	if err := cmd.WithDir(c.GetRepoPath()).Run(); err != nil {
		return nil, err, nil
	}

	commitRes := c.commitListRegexp.FindAllStringSubmatch(cmd.Stdout.String(), -1)
	tagList := make([]Commit, 0)
	branchList := make([]Commit, 0)
	commitList := make([]Commit, len(commitRes))
	for i, cr := range commitRes {
		makeCommit := func(cr []string, ref string) Commit {
			return Commit{
				Ref:    ref,
				Commit: cr[2],
				Author: cr[3],
				Date:   cr[4],
				Text:   cr[5],
			}
		}
		// Parse refs
		refRes := strings.Split(cr[1], ", ")
		for _, ref := range refRes {
			if strings.HasPrefix(ref, "tag: ") {
				// If commit has a tag ref add to tag list
				tagRef := fmt.Sprintf("tags/%s", strings.TrimPrefix(ref, "tag: "))
				commit := makeCommit(cr, tagRef)
				tagList = append(tagList, commit)
			} else if strings.HasPrefix(ref, "origin/") {
				// If commit has a head ref add to branch list
				branchRef := strings.TrimPrefix(ref, "origin/")
				commit := makeCommit(cr, branchRef)
				branchList = append(branchList, commit)
			}
		}
		// Add to commit list
		commitList[i] = makeCommit(cr, "")
	}
	return append(tagList, append(branchList, commitList...)...), nil, nil
}
