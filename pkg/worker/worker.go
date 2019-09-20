package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/makerdao/testchain-deployment/pkg/config"
	"github.com/makerdao/testchain-deployment/pkg/deploy"
	"github.com/makerdao/testchain-deployment/pkg/gateway"
	"github.com/makerdao/testchain-deployment/pkg/github"
	"github.com/makerdao/testchain-deployment/pkg/storage"
	gonats "github.com/nats-io/go-nats"
	"github.com/sirupsen/logrus"
)

type RunConfig struct {
	StepID        int
	RequestID     string
	DeployEnvVars map[string]string
}

type Worker struct {
	Storage         storage.Storage
	gatewayClient   *gateway.Client
	githubClient    *github.Client
	deployComponent *deploy.Component
}

func (w *Worker) Run(log *logrus.Entry) error {
	if err := w.deployComponent.FirstUpdate(log); err != nil {
		return err
	}
	runConfig, err := ParseEnvInput()
	if err != nil {
		return err
	}
	log.Debugf("Running deployment with config: %+v", runConfig)

	resultReq := &gateway.RunResultRequest{
		ID: runConfig.RequestID,
	}

	if resErr := w.deployComponent.RunStep(log, runConfig.StepID, runConfig.DeployEnvVars); resErr != nil {
		resultReq.Type = gateway.RunResultRequestTypeErr
		errResBytes, err := json.Marshal(resErr)
		if err != nil {
			log.WithError(err).Error("Can't marshal error for run result")
			return err
		}
		resultReq.Result = errResBytes
		if err := w.gatewayClient.RunResult(log, resultReq); err != nil {
			log.WithError(err).Error("Can't send request with result of run to gateway with error")
			return err
		}
		log.Errorf(resErr.Msg)
		return fmt.Errorf("failed to run deployment with error %+v", resErr)
	}

	log.Debugf("Finished deployment step")

	res, resErr := w.deployComponent.ReadResult()
	if resErr != nil {
		resultReq.Type = gateway.RunResultRequestTypeErr
		errResBytes, err := json.Marshal(resErr)
		if err != nil {
			log.WithError(err).Error("Can't marshal error for read result")
		}
		resultReq.Result = errResBytes
		if err := w.gatewayClient.RunResult(log, resultReq); err != nil {
			log.WithError(err).Error("Can't send request with result of run to gateway with error")
		}
		log.Errorf(resErr.Msg)
		return fmt.Errorf("failed to read deployment result with error %+v", resErr)
	}

	resBytes, err := json.Marshal(res)
	if err != nil {
		log.WithError(err).Error("Can't marshal error for run result")
	}
	resultReq.Type = gateway.RunResultRequestTypeOK
	resultReq.Result = resBytes
	if err := w.gatewayClient.RunResult(log, resultReq); err != nil {
		log.WithError(err).Error("Can't send request with result of run to gateway")
		return err
	}
	return nil
}

func (w *Worker) Shutdown(ctx context.Context, log *logrus.Entry) error {
	log.Debug("Start graceful shutdown of worker")
	defer log.Debug("Graceful shutdown of worker: done")
	//wait while all operations will be finished
	opCh := make(chan struct{})
	go func() {
		for {
			if !w.Storage.GetRun() {
				opCh <- struct{}{}
				return
			}
			time.Sleep(10 * time.Second)
		}
	}()
	select {
	case <-opCh:
		// TODO: check error
		return nil
	case <-ctx.Done():
		//return fmt.Errorf("context cancalled, but operations not colmpleted, server err: %w", err.Error())
		return nil
	}
}

func ParseEnvInput() (*RunConfig, error) {
	stepIDStr := os.Getenv("STEP_ID")
	stepID, err := strconv.Atoi(stepIDStr)
	if err != nil {
		return nil, err
	}
	if stepID == 0 {
		return nil, fmt.Errorf("wrong STEP_ID passed for deployment %d", stepID)
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
		StepID:        stepID,
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
	githubClient := github.NewClient(cfg.Github, cfg.Deploy.DeploymentDirPath)
	inMemStorage := storage.NewInMemory()
	deployComponent := deploy.New(cfg.Deploy, githubClient, inMemStorage)

	worker := &Worker{
		Storage:         inMemStorage,
		gatewayClient:   gatewayClient,
		githubClient:    githubClient,
		deployComponent: deployComponent,
	}

	return worker.Run(log)
}
