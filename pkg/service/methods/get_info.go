package methods

import (
	"encoding/json"

	"github.com/makerdao/testchain-deployment/pkg/deploy"
	"github.com/makerdao/testchain-deployment/pkg/serror"
	"github.com/sirupsen/logrus"
)

//GetInfoResponse response struct
type GetInfoResponse struct {
	Steps   []deploy.StepModel `json:"steps"`
	TagHash string             `json:"tagHash"`
}

//GetInfo return info about steps and commit's hash of tag
func (m *Methods) GetInfo(
	log *logrus.Entry,
	ID string,
	requestBytes []byte,
) (response []byte, error *serror.Error) {
	if !m.storage.HasData() {
		return nil, serror.New(serror.ErrCodeNotFound, "Has not loaded data")
	}
	hash, err := m.storage.GetTagHash(log)
	if err != nil {
		return nil, serror.New(serror.ErrCodeInternalError, "Can't get commit hash for tag")
	}

	stepList, err := m.storage.GetStepList(log)
	if err != nil {
		return nil, serror.New(serror.ErrCodeInternalError, "Can't get step list")
	}

	resp := GetInfoResponse{
		Steps:   stepList,
		TagHash: hash,
	}

	respBytes, err := json.Marshal(resp)
	if err != nil {
		return nil, serror.NewMarshalRespErr(err)
	}
	return respBytes, nil
}
