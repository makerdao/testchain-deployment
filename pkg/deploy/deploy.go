package deploy

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/sirupsen/logrus"

	"github.com/makerdao/testchain-deployment/pkg/command"
	"github.com/makerdao/testchain-deployment/pkg/github"
)

// StorageInterface for deploy action
type StorageInterface interface {
	UpsertManifest(log *logrus.Entry, manifest Manifest) error
	GetManifest(log *logrus.Entry) (*Manifest, error)
	GetScenario(log *logrus.Entry, scenarioNr int) (*Scenario, error)
	HasScenario(log *logrus.Entry, scenarioNr int) (bool, error)
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

//GetManifest return deploy manifest with available scenarios
func (c *Component) GetManifest(log *logrus.Entry) (*Manifest, error) {
	return c.storage.GetManifest(log)
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

//Checkout to commit
func (c *Component) Checkout(log *logrus.Entry, commit string) error {
	if err := c.storage.SetUpdate(true); err != nil {
		return err
	}
	defer func() {
		if err := c.storage.SetUpdate(false); err != nil {
			log.WithError(err).Error("Can't reset update status")
		}
	}()

	if cmdErr := c.githubClient.CheckoutCmd(log, &commit); cmdErr != nil {
		log.WithError(cmdErr).Error("Can't checkout to commit")
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

	return c.CollectInfo(log)
}

func (c *Component) FirstUpdate(log *logrus.Entry) error {
	log.Info(c.cfg.RunUpdateOnStart)
	switch c.cfg.RunUpdateOnStart {
	case "disable":
		return c.CollectInfo(log)
	case "enable":
		// first load source
		log.Info("First update src started, it takes a few minutes")
		if err := c.UpdateSource(log); err != nil {
			log.WithError(err).Error("Can't first update source")
			return err
		}
		log.Info("First update src finished")
		return nil
	case "ifNotExists":
		if err := c.MkDeploymentDirIfNotExists(log); err != nil {
			log.WithError(err).Error("Can't create dir for deployment")
			return err
		}
		empty, err := isDirEmpty(c.cfg.DeploymentDirPath)
		if err != nil {
			return err
		}
		log.Infof("Deployment dir exists: %+v", empty)
		if !empty {
			return c.CollectInfo(log)
		}
		// first load source
		log.Info("First update src started, it takes a few minutes")
		if err := c.UpdateSource(log); err != nil {
			log.WithError(err).Error("Can't first update source")
			return err
		}
		log.Info("First update src finished")
		return nil
	default:
		return errors.New("unknown strategy for update on start")
	}
}

func isDirEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	// read in ONLY one file
	_, err = f.Readdir(1)

	// and if the file is EOF... well, the dir is empty.
	if err == io.EOF {
		return true, nil
	}
	return false, err
}

func (c *Component) CollectInfo(log *logrus.Entry) error {
	tagHash, cmdErr := c.githubClient.LastHashCommitCmd(log)
	if cmdErr != nil {
		return cmdErr
	}

	manifest, err := c.getManifest(log)
	if err == nil {
		if err := c.storage.UpsertManifest(log, *manifest); err != nil {
			return err
		}
		log.Debugf("Loaded manifest:\n\n%+v", manifest)
	} else {
		log.Debugf("Couldn't read manifest file: %s", err)
	}

	if err := c.storage.SetTagHash(log, tagHash); err != nil {
		return err
	}

	if err := c.storage.SetUpdatedAtNow(); err != nil {
		return err
	}
	log.Debugf("Loaded hash: %+v", tagHash)

	return nil
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
	if cmdErr := c.githubClient.CheckoutCmd(log, nil); cmdErr != nil {
		log.WithError(cmdErr).Error("Can't checkout to tag")
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

	return c.CollectInfo(log)
}

// RunScenario run step command
// TODO run with ctx, and ability for stop
func (c *Component) RunScenario(log *logrus.Entry, scenarioNr int, envVars map[string]string) *ResultErrorModel {
	if err := c.storage.SetRun(true); err != nil {
		return NewResultErrorModelFromErr(err)
	}
	defer func() {
		if err := c.storage.SetRun(false); err != nil {
			log.WithError(err).Error("Can't reset run status")
		}
	}()

	scenario, err := c.storage.GetScenario(log, scenarioNr)
	if err != nil {
		return NewResultErrorModelFromErr(err)
	}
	cmd := command.New(exec.Command("nix", "run",
		"-f", ".", // Use Nix expression from current working directory for now
		"-c", scenario.RunCommand)).
		WithDir(c.githubClient.GetRepoPath()).
		WithEnvVarsMap(envVars)
	if cmdErr := cmd.Run(); cmdErr != nil {
		log.WithError(cmdErr.Message).Error("Cmd running error")
		log.Debugf("Cmd running error trace: %s", string(cmdErr.Stderr))
		return NewResultErrorModelFromCmd(cmdErr)
	}
	return nil
}

//ReadResult from configured file
func (c *Component) ReadResult() (*ResultModel, *ResultErrorModel) {
	fileName := filepath.Join(c.githubClient.GetRepoPath(), c.cfg.ResultSubPath)
	fi, err := os.Stat(fileName)
	if err != nil {
		return nil, NewResultErrorModelFromErr(err)
	}
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, NewResultErrorModelFromErr(err)
	}
	return NewResultModel(fi.ModTime(), data), nil
}

func (c *Component) GetCommitList(log *logrus.Entry) ([]github.Commit, *ResultErrorModel) {
	res, cmdErr, err := c.githubClient.GetCommitList(log)
	if err != nil {
		return nil, NewResultErrorModelFromErr(err)
	}
	if cmdErr != nil {
		return nil, NewResultErrorModelFromCmd(cmdErr)
	}
	return res, nil
}

func (c *Component) getManifest(log *logrus.Entry) (*Manifest, error) {
	dirPath := c.githubClient.GetRepoPath()
	log.Debugf("Read manifest from repo: %s", dirPath)
	model, err := ReadManifestFile(ioutil.ReadFile, dirPath)
	if err != nil {
		return nil, err
	}
	return model, nil
}

type ReadFile func(path string) ([]byte, error)

func ReadManifestFile(readFile ReadFile, repoPath string) (*Manifest, error) {
	path := filepath.Join(repoPath, ".staxx-scenarios")
	data, err := readFile(path)
	if err != nil {
		return nil, err
	}
	var model ManifestModel
	if err := json.Unmarshal(data, &model); err != nil {
		return nil, err
	}

	scenarios := make([]Scenario, len(model.Scenarios))
	for i, scenario := range model.Scenarios {
		configPath := filepath.Join(repoPath, scenario.ConfigPath)
		config, err := readFile(configPath)
		if err != nil {
			return nil, err
		}
		var configModel json.RawMessage
		if err := json.Unmarshal(config, &configModel); err != nil {
			return nil, err
		}
		scenarios[i] = Scenario{
			scenario.Name,
			scenario.Description,
			scenario.RunCommand,
			configModel,
			scenario.OutPath,
		}
	}

	manifest := Manifest{
		model.Name,
		model.Description,
		scenarios,
	}
	return &manifest, nil
}
