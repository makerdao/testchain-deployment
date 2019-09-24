package methods

import (
	"encoding/json"
	"fmt"

	"github.com/makerdao/testchain-deployment/pkg/git"
	"github.com/makerdao/testchain-deployment/pkg/serror"
	"github.com/sirupsen/logrus"
)

// GetRefsRequest request data
type GetRefsRequest struct {
	URL string `json:"url"`
}

// Return remote refs for a GIT repo URL
func (m *Methods) GetRefs(
	log *logrus.Entry,
	id string,
	requestBytes []byte,
) (response []byte, error *serror.Error) {
	var req GetRefsRequest
	var res []git.Commit
	if err := json.Unmarshal(requestBytes, &req); err != nil {
		return nil, serror.NewUnmarshalReqErr(err)
	}

	res, resErr := git.GetRefs(req.URL)
	if resErr != nil {
		return nil, serror.New(serror.ErrCodeInternalError,
			fmt.Sprintf("Couldn't get refs for: %s", req.URL),
			resErr)
	}

	resBytes, err := json.Marshal(res)
	if err != nil {
		return nil, serror.NewMarshalRespErr(err)
	}

	return resBytes, nil
}
