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
		return nil, serror.New(serror.ErrCodeBadRequest, "Update source already run")
	}

	go func(id string) {
		resultReq := &gateway.UpdateResultRequest{
			ID: id,
		}
		if err := m.storage.SetUpdate(true); err != nil {
			if err := m.gatewayClient.UpdateResult(log, resultReq.SetErr(err)); err != nil {
				log.WithError(err).Error("Can't send request with result of run to gateway with error")
			}
			return
		}
		defer func() {
			if err := m.storage.SetUpdate(false); err != nil {
				log.WithError(err).Error("Can't reset update status")
			}
		}()
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
