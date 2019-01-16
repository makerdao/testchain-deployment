package deploy

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/makerdao/testchain-deployment/pkg/github"
	"github.com/mholt/archiver"
	"github.com/sirupsen/logrus"
)

// StorageInterface for deploy action
type StorageInterface interface {
	UpsertStepList(log *logrus.Entry, modelList []Model) error
	GetStepList(log *logrus.Entry) ([]Model, error)
	HasStepList(log *logrus.Entry, id int) (bool, error)
	SetTagHash(log *logrus.Entry, hash string) error
	GetTagHash(log *logrus.Entry) (hash string, err error)
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

//GetStepList return list of available steps
func (c *Component) GetStepList(log *logrus.Entry) ([]Model, error) {
	return c.storage.GetStepList(log)
}

//GetTagHash return hash commit of tag
func (c *Component) GetTagHash(log *logrus.Entry) (string, error) {
	return c.storage.GetTagHash(log)
}

//RemoveSourceDirIfExists remove source dir if exist
func (c *Component) RemoveSourceDirIfExists() error {
	path := c.getStepListFilePath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}
	if err := os.RemoveAll(path); err != nil {
		return err
	}
	return nil
}

//CleanData clean all external data(github data and deployment data)
func (c *Component) CleanData() error {
	if err := c.RemoveSourceDirIfExists(); err != nil {
		return err
	}
	if err := c.githubClient.RemoveArchiveIfExists(); err != nil {
		return err
	}
	return nil
}

//UpdateSource from github and prepare for work with it
func (c *Component) UpdateSource(log *logrus.Entry) error {
	if err := c.CleanData(); err != nil {
		return err
	}
	tagHash, err := c.githubClient.GetTagHash(log.WithField("component", "gitClient"))
	if err != nil {
		return err
	}
	filePath, err := c.githubClient.DownloadTarGzSourceCode(log)
	if err != nil {
		return err
	}
	tar := archiver.NewTarGz()
	tar.ImplicitTopLevelFolder = false
	if err := tar.Unarchive(filePath, c.cfg.DeploymentDirPath); err != nil {
		return err
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

	log.Debugf("Loaded data: \n %+v \n\n %+v", tagHash, stepList)

	return nil
}

// RunStep run step command
// TODO run with ctx, and ability for stop
func (c *Component) RunStep(log *logrus.Entry, stepID int) error {
	hasStep, err := c.storage.HasStepList(log, stepID)
	if err != nil {
		return err
	}
	if !hasStep {
		return errors.New("unknown id of step")
	}
	commandName := fmt.Sprintf("./step-%d-deploy", stepID)
	cmd := exec.Command(commandName)
	cmd.Dir = c.getStepListFilePath()
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func (c *Component) getStepList(log *logrus.Entry) ([]Model, error) {
	stepList := make([]Model, 0)
	dirPath := c.getStepListFilePath()
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

func readStepDescriptionFile(path string) (*Model, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var model Model
	if err := json.Unmarshal(bytes, &model); err != nil {
		return nil, err
	}
	return &model, nil
}

func (c *Component) getStepListFilePath() string {
	return filepath.Join(c.cfg.DeploymentDirPath, c.githubClient.GetDirInArchiveFromCfg(), c.cfg.DeploymentSubPath)
}
