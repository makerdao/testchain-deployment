package git

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

const (
	nixExpr = `toString (fetchGit {
    url = "%s";
    ref = "%s";
    rev = "%s";
  })`
)

type Commit struct {
	URL string `json:"url"`
	Ref string `json:"ref"`
	Rev string `json:"rev"`
}

func commitToNix(commit Commit) string {
	return fmt.Sprintf(nixExpr, commit.URL, commit.Ref, commit.Rev)
}

func runCmd(cmd *exec.Cmd) (string, error) {
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("Command failed: %s: %+v\nSTDERR: %s",
				strings.Join(cmd.Args, " "),
				exitErr,
				exitErr.Stderr)
		}
		return "", fmt.Errorf("Command failed: %s: %+v",
			strings.Join(cmd.Args, " "),
			err)
	}
	return string(out), nil
}

func GetRefs(url string) ([]Commit, error) {
	remoteRefParser := regexp.MustCompile(`(?m)^(.*)[ \t]+(.*)$`)

	stdout, err := runCmd(exec.Command("git", "ls-remote", url))
	if err != nil {
		return nil, err
	}

	remoteRes := remoteRefParser.FindAllStringSubmatch(stdout, -1)
	refList := make([]Commit, len(remoteRes))

	for i, cr := range remoteRes {
		refList[i] = Commit{
			URL: url,
			Rev: cr[1],
			Ref: cr[2],
		}
	}
	return refList, nil
}

func GetRepoPath(commit Commit) (string, error) {
	stdout, err := runCmd(exec.Command("nix-instantiate", "--eval", "--json", "-E", commitToNix(commit)))
	if err != nil {
		return "", fmt.Errorf("Failed to checkout GIT repo %s %s: %+v", commit.URL, commit.Rev, err)
	}
	if stdout == "" {
		return "", fmt.Errorf("Failed to get path to repo %s %s", commit.URL, commit.Rev)
	}
	return strings.Trim(stdout, `"`), nil
}
