package deploy

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/makerdao/testchain-deployment/pkg/command"
	"github.com/makerdao/testchain-deployment/pkg/dapp"
	"github.com/makerdao/testchain-deployment/pkg/github"
	"github.com/sirupsen/logrus"
)

// StorageInterface for deploy action
type StorageInterface interface {
	UpsertStepList(log *logrus.Entry, modelList []StepModel) error
	GetStepList(log *logrus.Entry) ([]StepModel, error)
	HasStep(log *logrus.Entry, id int) (bool, error)
	SetTagHash(log *logrus.Entry, hash string) error
	GetTagHash(log *logrus.Entry) (hash string, err error)
	SetUpdate(run bool) error
	GetUpdate() bool
	SetRun(run bool) error
	GetRun() bool
	SetUpdatedAtNow() error
}

// Component is main module of deploy
type Component struct {
	cfg            Config
	githubClient   *github.Client
	dappClient     *dapp.Client
	stepNameRegexp *regexp.Regexp
	storage        StorageInterface
}

// New init component
func New(cfg Config, githubClient *github.Client, storage StorageInterface) *Component {
	return &Component{
		cfg:            cfg,
		githubClient:   githubClient,
		stepNameRegexp: regexp.MustCompile(`^step-(\d+)\.json$`),
		storage:        storage,
	}
}

//GetStepList return list of available steps
func (c *Component) GetStepList(log *logrus.Entry) ([]StepModel, error) {
	return c.storage.GetStepList(log)
}

//GetTagHash return hash commit of tag
func (c *Component) GetTagHash(log *logrus.Entry) (string, error) {
	return c.storage.GetTagHash(log)
}

// MkDeploymentDirIfNotExists create base dir for work if not exists
func (c *Component) MkDeploymentDirIfNotExists(log *logrus.Entry) error {
	_, err := os.Stat(c.cfg.DeploymentDirPath)
	if os.IsExist(err) {
		return nil
	}
	return os.MkdirAll(c.cfg.DeploymentDirPath, 0755)
}

//UpdateSource from github and prepare for work with it
func (c *Component) UpdateSource(log *logrus.Entry) error {
	if err := c.storage.SetUpdate(true); err != nil {
		return err
	}
	defer func() {
		if err := c.storage.SetUpdate(false); err != nil {
			log.WithError(err).Error("Can't reset update status")
		}
	}()

	if err := c.MkDeploymentDirIfNotExists(log); err != nil {
		log.WithError(err).Error("Can't create dir for deployment")
		return err
	}
	if err := c.githubClient.CleanLoadingRepoIfExists(log); err != nil {
		log.WithError(err).Error("Can't clean deployment dir")
		return err
	}
	if cmdErr := c.githubClient.CloneCmd(log); cmdErr != nil {
		log.WithError(cmdErr).Error("Can't clone deployment")
		return cmdErr
	}
	if cmdErr := c.githubClient.UpdateSubmodulesCmd(log); cmdErr != nil {
		log.WithError(cmdErr).Error("Can't update submodules deployment")
		return cmdErr
	}
	if cmdErr := c.githubClient.CheckoutCmd(log, nil); cmdErr != nil {
		log.WithError(cmdErr).Error("Can't checkout to tag")
		return cmdErr
	}
	if cmdErr := c.dappClient.UpdateCmd(c.githubClient.GetLoadingPath()); cmdErr != nil {
		log.WithError(cmdErr).Error("Can't dapp update")
		return cmdErr
	}

	// if we have error while downloading repo, we can work with prev version of repo
	if err := c.githubClient.CleanRepoIfExists(log); err != nil {
		log.WithError(err).Error("Can't clean deployment dir")
		return err
	}
	cmdErr := command.New(
		exec.Command("cp", "-r", c.githubClient.GetLoadingPath(), c.githubClient.GetRepoPath()),
	).Run()
	if cmdErr != nil {
		return cmdErr
	}

	tagHash, cmdErr := c.githubClient.LastHashCommitCmd(log)
	if cmdErr != nil {
		return cmdErr
	}

	stepList, err := c.getStepList(log)
	if err != nil {
		return err
	}

	if err := c.storage.UpsertStepList(log, stepList); err != nil {
		return err
	}

	if err := c.storage.SetTagHash(log, tagHash); err != nil {
		return err
	}

	if err := c.storage.SetUpdatedAtNow(); err != nil {
		return err
	}

	log.Debugf("Loaded data: \n %+v \n\n %+v", tagHash, stepList)

	return nil
}

// RunStep run step command
// TODO run with ctx, and ability for stop
func (c *Component) RunStep(log *logrus.Entry, stepID int, envVars map[string]string) *ResultErrorModel {
	if err := c.storage.SetRun(true); err != nil {
		return NewResultErrorModelFromErr(err)
	}
	defer func() {
		if err := c.storage.SetRun(false); err != nil {
			log.WithError(err).Error("Can't reset run status")
		}
	}()

	hasStep, err := c.storage.HasStep(log, stepID)
	if err != nil {
		return NewResultErrorModelFromErr(err)
	}
	if !hasStep {
		return NewResultErrorModelFromTxt("unknown id of step")
	}
	cmd := command.New(exec.Command(fmt.Sprintf("./step-%d-deploy", stepID))).
		WithDir(c.githubClient.GetRepoPath()).
		WithEnvVarsMap(envVars)
	if cmdErr := cmd.Run(); cmdErr != nil {
		log.WithError(cmdErr.Message).Error("Cmd running error")
		log.Debugf("Cmd running error trace: %s", string(cmdErr.Stderr))
		return NewResultErrorModelFromCmd(cmdErr)
	}
	return nil
}

func (c *Component) getStepList(log *logrus.Entry) ([]StepModel, error) {
	stepList := make([]StepModel, 0)
	dirPath := c.githubClient.GetRepoPath()
	log.Debugf("Read step list from: %s", dirPath)
	wErr := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == dirPath {
			return nil
		}
		if info.IsDir() {
			return filepath.SkipDir
		}
		res := c.stepNameRegexp.FindStringSubmatch(info.Name())
		if len(res) != 2 {
			return nil
		}
		step, err := strconv.Atoi(res[1])
		if err != nil {
			return err
		}
		model, err := readStepDescriptionFile(path)
		if err != nil {
			return err
		}
		model.ID = step
		stepList = append(stepList, *model)
		return nil
	})
	if wErr != nil {
		return nil, wErr
	}
	return stepList, nil
}

func readStepDescriptionFile(path string) (*StepModel, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var model StepModel
	if err := json.Unmarshal(bytes, &model); err != nil {
		return nil, err
	}
	return &model, nil
}
