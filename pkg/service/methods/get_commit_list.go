package methods

import (
	"github.com/sirupsen/logrus"

	"github.com/makerdao/testchain-deployment/pkg/gateway"
	"github.com/makerdao/testchain-deployment/pkg/serror"
)

//GetCommitList source if it possible
func (m *Methods) GetCommitList(
	log *logrus.Entry,
	id string,
	requestBytes []byte,
) (response []byte, error *serror.Error) {
	if m.c {
		return nil, serror.New(serror.ErrCodeInternalError, "Deploy script updating in progress")
	}
	if m.storage.GetRun() {
		return nil, serror.New(serror.ErrCodeInternalError, "Deploy script running in progress")
	}
	go func(id string) {
		resultReq := &gateway.UpdateResultRequest{
			ID: id,
		}
		if err := m.deployComponent.UpdateSource(log); err != nil {
			if err := m.gatewayClient.UpdateResult(log, resultReq.SetErr(err)); err != nil {
				log.WithError(err).Error("Can't send request with result of run to gateway with error")
			}
			return
		}
		resultReq.Type = gateway.UpdateResultRequestTypeOK
		if err := m.gatewayClient.UpdateResult(log, resultReq); err != nil {
			log.WithError(err).Error("Can't send request with result of run to gateway")
		}
	}(id)

	return []byte(`{}`), nil
}
