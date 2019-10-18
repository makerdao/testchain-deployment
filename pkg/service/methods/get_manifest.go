package methods

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/makerdao/testchain-deployment/pkg/deploy"
	"github.com/makerdao/testchain-deployment/pkg/git"
	"github.com/makerdao/testchain-deployment/pkg/serror"
	"github.com/sirupsen/logrus"
)

// GetManifestReq request data
type GetManifestReq = git.Commit

// Return remote refs for a GIT repo URL
func (m *Methods) GetManifest(
	log *logrus.Entry,
	id string,
	requestBytes []byte,
) (response []byte, error *serror.Error) {
	var req GetManifestReq
	if err := json.Unmarshal(requestBytes, &req); err != nil {
		return nil, serror.NewUnmarshalReqErr(err)
	}

	repoPath, repoErr := git.GetRepoPath(req)
	if repoErr != nil {
		return nil, serror.New(serror.ErrCodeInternalError,
			fmt.Sprintf("Couldn't get path to repo: %s %s", req.URL, req.Rev),
			repoErr)
	}
	manifest, manifestErr := deploy.ReadManifestFile(ioutil.ReadFile, repoPath)
	if manifestErr != nil {
		return nil, serror.New(serror.ErrCodeInternalError,
			fmt.Sprintf("Couldn't get manifest file for: %s %s", req.URL, req.Rev),
			manifestErr)
	}

	resBytes, err := json.Marshal(manifest)
	if err != nil {
		return nil, serror.NewMarshalRespErr(err)
	}

	return resBytes, nil
}
