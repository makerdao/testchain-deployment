package methods

import (
	"encoding/json"

	"github.com/sirupsen/logrus"

	"github.com/makerdao/testchain-deployment/pkg/gateway"
	"github.com/makerdao/testchain-deployment/pkg/serror"
)

type CommitRequest struct {
	Commit string `json:"commit"`
}

//Checkout to commit if it possible
func (m *Methods) Checkout(
	log *logrus.Entry,
	id string,
	requestBytes []byte,
) (response []byte, error *serror.Error) {
	if m.storage.GetUpdate() {
		return nil, serror.New(serror.ErrCodeInternalError, "Deploy script updating in progress")
	}
	if m.storage.GetRun() {
		return nil, serror.New(serror.ErrCodeInternalError, "Deploy script running in progress")
	}
	var req CommitRequest
	if err := json.Unmarshal(requestBytes, &req); err != nil {
		return nil, serror.New(serror.ErrCodeBadRequest, "Can't decode request")
	}
	go func(id string) {
		resultReq := &gateway.CheckoutResultRequest{
			ID: id,
		}
		if err := m.deployComponent.Checkout(log, req.Commit); err != nil {
			if err := m.gatewayClient.CheckoutResult(log, resultReq.SetErr(err)); err != nil {
				log.WithError(err).Error("Can't send request with result of run to gateway with error")
			}
			return
		}
		resultReq.Type = gateway.CheckoutResultRequestTypeOK
		if err := m.gatewayClient.CheckoutResult(log, resultReq); err != nil {
			log.WithError(err).Error("Can't send request with result of run to gateway")
		}
	}(id)

	return []byte(`{}`), nil
}
