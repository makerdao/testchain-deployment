package methods

import (
	"github.com/makerdao/testchain-deployment/pkg/gateway"
	"github.com/makerdao/testchain-deployment/pkg/serror"
	"github.com/sirupsen/logrus"
)

//Update source if it possible
func (m *Methods) Update(
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

	log.Debugf("Update source process started with request Id %s", id)
	
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
