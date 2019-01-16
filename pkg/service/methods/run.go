package methods

import (
	"encoding/json"

	"github.com/makerdao/testchain-deployment/pkg/gateway"
	"github.com/makerdao/testchain-deployment/pkg/serror"
	"github.com/sirupsen/logrus"
)

//RunRequest request data
type RunRequest struct {
	StepID int `json:"stepId"`
}

//Run deployment async and return ok if it possible
func (m *Methods) Run(
	log *logrus.Entry,
	id string,
	requestBytes []byte,
) (response []byte, error *serror.Error) {
	var req RunRequest
	if err := json.Unmarshal(requestBytes, &req); err != nil {
		return nil, serror.NewUnmarshalReqErr(err)
	}

	if m.storage.GetRun() {
		return nil, serror.New(serror.ErrCodeBadRequest, "Deployment already run")
	}

	go func(id string, req RunRequest) {
		resultReq := &gateway.RunResultRequest{
			ID: id,
		}
		if err := m.storage.SetRun(true); err != nil {
			if err := m.gatewayClient.RunResult(log, resultReq.SetErr(err)); err != nil {
				log.WithError(err).Error("Can't send request with result of run to gateway with error")
			}
			return
		}
		defer func() {
			if err := m.storage.SetRun(false); err != nil {
				log.WithError(err).Error("Can't reset run status")
			}
		}()
		if err := m.deployComponent.RunStep(log, req.StepID); err != nil {
			if err := m.gatewayClient.RunResult(log, resultReq.SetErr(err)); err != nil {
				log.WithError(err).Error("Can't send request with result of run to gateway with error")
			}
			return
		}
		resultReq.Type = gateway.RunResultRequestTypeOK
		if err := m.gatewayClient.RunResult(log, resultReq); err != nil {
			log.WithError(err).Error("Can't send request with result of run to gateway")
		}
	}(id, req)

	return []byte(`{}`), nil
}
