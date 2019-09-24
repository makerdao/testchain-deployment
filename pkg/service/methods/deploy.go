package methods

import (
	"encoding/json"

	"github.com/makerdao/testchain-deployment/pkg/deploy"
	"github.com/makerdao/testchain-deployment/pkg/gateway"
	"github.com/makerdao/testchain-deployment/pkg/git"
	"github.com/makerdao/testchain-deployment/pkg/serror"
	"github.com/sirupsen/logrus"
)

//DeployRequest request data
type DeployRequest struct {
	RepoURL    string            `json:"repoUrl"`
	RepoRef    string            `json:"repoRef"`
	RepoRev    string            `json:"repoRev"`
	ScenarioNr int               `json:"scenarioNr"`
	EnvVars    map[string]string `json:"envVars"`
}

//Run deployment async and return ok if it possible
func (m *Methods) Deploy(
	log *logrus.Entry,
	id string,
	requestBytes []byte,
) (response []byte, error *serror.Error) {
	var req DeployRequest
	if err := json.Unmarshal(requestBytes, &req); err != nil {
		return nil, serror.NewUnmarshalReqErr(err)
	}

	go func(id string, req DeployRequest) {
		resultReq := &gateway.RunResultRequest{
			ID: id,
		}
		deployment := deploy.Deployment{
			Commit: git.Commit{
				URL: req.RepoURL,
				Ref: req.RepoRef,
				Rev: req.RepoRev,
			},
			ScenarioNr:    req.ScenarioNr,
			DeployEnvVars: req.EnvVars,
		}

		res, resErr := deploy.Deploy(log, deployment)
		if resErr != nil {
			resultReq.Type = gateway.RunResultRequestTypeErr
			errResBytes, err := json.Marshal(resErr)
			if err != nil {
				log.WithError(err).Error("Can't marshal error for deploy result")
			}
			resultReq.Result = errResBytes
			if err := m.gatewayClient.RunResult(log, resultReq); err != nil {
				log.WithError(err).Error("Can't send request with result of deployment to gateway with error")
			}
			return
		}

		resBytes, err := json.Marshal(res)
		if err != nil {
			log.WithError(err).Error("Can't marshal error for deployment result")
		}
		resultReq.Type = gateway.RunResultRequestTypeOK
		resultReq.Result = resBytes
		if err := m.gatewayClient.RunResult(log, resultReq); err != nil {
			log.WithError(err).Error("Can't send request with result of run to gateway")
		}
	}(id, req)

	return []byte(`{}`), nil
}
