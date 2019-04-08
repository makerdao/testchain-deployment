package methods

import (
	"encoding/json"

	"github.com/sirupsen/logrus"

	"github.com/makerdao/testchain-deployment/pkg/github"
	"github.com/makerdao/testchain-deployment/pkg/serror"
)

type GetCommitListResponse struct {
	Data []github.Commit `json:"data"`
}

//GetCommitList source if it possible
func (m *Methods) GetCommitList(
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
	res, dErr := m.deployComponent.GetCommitList(log)
	if dErr != nil {
		return nil, serror.New(serror.ErrCodeInternalError, "Can't get list of commits")
	}

	respBytes, err := json.Marshal(GetCommitListResponse{Data: res})
	if err != nil {
		return nil, serror.NewMarshalRespErr(err)
	}

	return respBytes, nil
}
