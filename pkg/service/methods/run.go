package methods

import (
	"encoding/json"

	"github.com/makerdao/testchain-deployment/pkg/gateway"
	"github.com/makerdao/testchain-deployment/pkg/serror"
	"github.com/sirupsen/logrus"
)

//RunRequest request data
type RunRequest struct {
	StepID  int               `json:"stepId"`
	EnvVars map[string]string `json:"envVars"`
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

	if m.storage.GetUpdate() {
		return nil, serror.New(serror.ErrCodeInternalError, "Deploy script updating in progress")
	}
	if m.storage.GetRun() {
		return nil, serror.New(serror.ErrCodeInternalError, "Deploy script running in progress")
	}

	go func(id string, req RunRequest) {
		resultReq := &gateway.RunResultRequest{
			ID: id,
		}
		if resErr := m.deployComponent.RunStep(log, req.StepID, req.EnvVars); resErr != nil {
			resultReq.Type = gateway.RunResultRequestTypeErr
			errResBytes, err := json.Marshal(resErr)
			if err != nil {
				log.WithError(err).Error("Can't marshal error for run result")
			}
			resultReq.Result = errResBytes
			if err := m.gatewayClient.RunResult(log, resultReq); err != nil {
				log.WithError(err).Error("Can't send request with result of run to gateway with error")
			}
			return
		}
		res, resErr := m.deployComponent.ReadResult()
		if resErr != nil {
			resultReq.Type = gateway.RunResultRequestTypeErr
			errResBytes, err := json.Marshal(resErr)
			if err != nil {
				log.WithError(err).Error("Can't marshal error for read result")
			}
			resultReq.Result = errResBytes
			if err := m.gatewayClient.RunResult(log, resultReq); err != nil {
				log.WithError(err).Error("Can't send request with result of run to gateway with error")
			}
			return
		}
		resBytes, err := json.Marshal(res)
		if err != nil {
			log.WithError(err).Error("Can't marshal error for run result")
		}
		resultReq.Type = gateway.RunResultRequestTypeOK
		resultReq.Result = resBytes
		if err := m.gatewayClient.RunResult(log, resultReq); err != nil {
			log.WithError(err).Error("Can't send request with result of run to gateway")
		}
	}(id, req)

	return []byte(`{}`), nil
}
