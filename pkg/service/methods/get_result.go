package methods

import (
	"encoding/json"

	"github.com/makerdao/testchain-deployment/pkg/serror"
	"github.com/sirupsen/logrus"
)

//GetInfo return info about steps and commit's hash of tag
func (m *Methods) GetResult(
	log *logrus.Entry,
	ID string,
	requestBytes []byte,
) (response []byte, error *serror.Error) {
	if !m.storage.HasData() {
		return nil, serror.New(serror.ErrCodeNotFound, "Has not loaded data")
	}
	if m.storage.GetRun() {
		return nil, serror.New(serror.ErrCodeNotFound, "Deployment is running now")
	}
	if m.storage.GetUpdate() {
		return nil, serror.New(serror.ErrCodeNotFound, "Deployment is updating now")
	}
	resp, appErr := m.deployComponent.ReadResult()
	if appErr != nil {
		return nil, serror.New(serror.ErrCodeInternalError, "Can't read result")
	}

	respBytes, err := json.Marshal(resp)
	if err != nil {
		return nil, serror.NewMarshalRespErr(err)
	}
	return respBytes, nil
}
