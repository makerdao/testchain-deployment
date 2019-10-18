package worker

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/makerdao/testchain-deployment/pkg/config"
	"github.com/makerdao/testchain-deployment/pkg/deploy"
	"github.com/makerdao/testchain-deployment/pkg/gateway"
	"github.com/makerdao/testchain-deployment/pkg/git"
	gonats "github.com/nats-io/go-nats"
	"github.com/sirupsen/logrus"
)

type RunConfig struct {
	RepoURL       string
	RepoRef       string
	RepoRev       string
	ScenarioNr    int
	RequestID     string
	DeployEnvVars map[string]string
}

type Worker struct {
	gatewayClient *gateway.Client
	log           *logrus.Entry
}

func (w *Worker) failJob(reqID string, resErr *deploy.ResultErrorModel) error {
	w.log.Errorf(resErr.Msg)

	errResBytes, err := json.Marshal(resErr)
	if err != nil {
		w.log.WithError(err).Error("Can't marshal error")
		return err
	}

	resultReq := &gateway.RunResultRequest{
		ID:     reqID,
		Type:   gateway.RunResultRequestTypeErr,
		Result: errResBytes,
	}
	if err := w.gatewayClient.RunResult(w.log, resultReq); err != nil {
		w.log.WithError(err).Error("Can't send error result to gateway")
		return err
	}

	return fmt.Errorf("Worker failed to run deployment with error %+v", resErr)
}

func (w *Worker) returnResult(reqID string, res *deploy.ResultModel) error {
	resBytes, err := json.Marshal(res)
	if err != nil {
		w.log.WithError(err).Error("Can't marshal error for run result")
		return w.failJob(reqID, deploy.NewResultErrorModelFromErr(err))
	}

	resultReq := &gateway.RunResultRequest{
		ID:     reqID,
		Type:   gateway.RunResultRequestTypeOK,
		Result: resBytes,
	}
	if err := w.gatewayClient.RunResult(w.log, resultReq); err != nil {
		w.log.WithError(err).Error("Can't send request with result of run to gateway")
		return err
	}

	return nil
}

func readResult(res json.RawMessage) *deploy.ResultModel {
	return deploy.NewResultModel(time.Now(), res)
}

func (w *Worker) Run() error {
	runConfig, err := ParseEnvInput()
	if err != nil {
		return err
	}
	deployment := deploy.Deployment{
		Commit: git.Commit{
			URL: runConfig.RepoURL,
			Ref: runConfig.RepoRef,
			Rev: runConfig.RepoRev,
		},
		ScenarioNr:    runConfig.ScenarioNr,
		DeployEnvVars: runConfig.DeployEnvVars,
	}

	res, err := deploy.Deploy(w.log, deployment)
	if err != nil {
		return w.failJob(runConfig.RequestID, deploy.NewResultErrorModelFromErr(err))
	}

	return w.returnResult(runConfig.RequestID, readResult(res))
}

func ParseEnvInput() (*RunConfig, error) {
	repoURL := os.Getenv("REPO_URL")
	repoRef := os.Getenv("REPO_REF")
	repoRev := os.Getenv("REPO_REV")
	if repoURL == "" {
		return nil, fmt.Errorf("Need to specify REPO_URL and optinally REPO_REF and REPO_REV")
	}

	scenarioNrStr := os.Getenv("SCENARIO_NR")
	scenarioNr, err := strconv.Atoi(scenarioNrStr)
	if err != nil {
		return nil, err
	}

	requestID := os.Getenv("REQUEST_ID")
	if requestID == "" {
		return nil, fmt.Errorf("wrong REQUEST_ID passed for deployment, %s", requestID)
	}

	deployEnvs := os.Getenv("DEPLOY_ENV")
	var envVars map[string]string
	if err := json.Unmarshal([]byte(deployEnvs), &envVars); err != nil {
		return nil, err
	}

	return &RunConfig{
		RepoURL:       repoURL,
		RepoRef:       repoRef,
		RepoRev:       repoRev,
		ScenarioNr:    scenarioNr,
		RequestID:     requestID,
		DeployEnvVars: envVars,
	}, nil
}

func Execute(log *logrus.Entry, cfg *config.Config) error {
	// init components
	natsConn, err := gonats.Connect(
		cfg.NATS.Servers,
		gonats.MaxReconnects(cfg.NATS.MaxReconnect),
		gonats.ReconnectWait(time.Duration(cfg.NATS.ReconnectWaitSec)*time.Second),
	)
	if err != nil {
		return err
	}

	gatewayClient := gateway.NewClient(cfg.Gateway, natsConn, cfg.NATS)

	worker := &Worker{
		gatewayClient: gatewayClient,
		log:           log,
	}

	return worker.Run()
}
