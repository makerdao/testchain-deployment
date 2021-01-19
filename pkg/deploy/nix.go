package deploy

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/makerdao/testchain-deployment/pkg/command"
	"github.com/makerdao/testchain-deployment/pkg/git"
	"github.com/sirupsen/logrus"
)

type Deployment struct {
	Commit        git.Commit
	ScenarioNr    int
	DeployEnvVars map[string]string
}

func Deploy(log *logrus.Entry, deployment Deployment) ([]byte, error) {
	log.Debugf("Starting deployment with: %+v", deployment)

	log.Debugf("Fetching GIT repo: %+v", deployment.Commit)
	repoPath, err := git.GetRepoPath(deployment.Commit)
	if err != nil {
		log.WithError(err).Error("Couldn't get repository")
		return nil, err
	}

	log.Debugf("Reading manifest file from: %s", repoPath)
	manifest, err := ReadManifestFile(ioutil.ReadFile, repoPath)
	if err != nil {
		log.WithError(err).Error("Couldn't read deploy manifest")
		return nil, err
	}
	if deployment.ScenarioNr < 0 || deployment.ScenarioNr >= len(manifest.Scenarios) {
		err := fmt.Errorf("Scenario number %d not available", deployment.ScenarioNr)
		log.WithError(err)
		return nil, err
	}
	scenario := manifest.Scenarios[deployment.ScenarioNr]

	workDir, err := ioutil.TempDir("", "deploy-worker-")
	if err != nil {
		log.WithError(err).Error("Couldn't create working directory")
		return nil, err
	}
	defer os.RemoveAll(workDir)

	log.Debugf("Working directory: %s", workDir)
	log.Debugf("Environment variables: %v", deployment.DeployEnvVars)
	log.Debugf("Running deployment command: %s", scenario.RunCommand)

	// Building list of arguments for `nix` command.
	args := []string{
		"run",
		"-f", repoPath,
		"-c",
	}
	args = append(args, strings.Split(scenario.RunCommand, " ")...)
	cmd := command.New(exec.Command("nix", args...)).
		WithDir(workDir).
		WithEnvVarsMap(deployment.DeployEnvVars)
	if cmdErr := cmd.Run(); cmdErr != nil {
		log.WithError(cmdErr.Message).
			Errorf("Error when running command: %s: %+v\nSTDERR: %s",
				strings.Join(cmd.Args, " "),
				cmdErr,
				string(cmdErr.Stderr))
		return nil, cmdErr
	}

	log.Debugf("Finished deploy command")

	outPath := filepath.Join(workDir, scenario.OutPath)

	log.Debugf("Reading deploy output from: %s", outPath)
	res, err := ioutil.ReadFile(outPath)
	if err != nil {
		return nil, err
	}

	return res, nil
}
