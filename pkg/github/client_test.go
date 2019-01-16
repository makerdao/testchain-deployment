package github

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestClient_DownloadTarGzSourceCode(t *testing.T) {
	r := require.New(t)
	cfg := GetDefaultConfig()
	r.Nil(cfg.Decode(os.Getenv("TCD_GITHUB")))
	r.Nil(cfg.Validate())
	logger := logrus.WithField("app", "test")

	client := NewClient(cfg)
	filePath, err := client.DownloadTarGzSourceCode(logger)
	r.Nil(err)
	r.Nil(filePath)
	r.Nil(os.Stat(filePath))
}

func TestClient_GetTagHash(t *testing.T) {
	r := require.New(t)
	cfg := GetDefaultConfig()
	r.Nil(cfg.Decode(os.Getenv("TCD_GITHUB")))
	r.Nil(cfg.Validate())
	logger := logrus.WithField("app", "test")

	client := NewClient(cfg)
	tagHas, err := client.GetTagHash(logger)
	r.Nil(err)
	r.True(len(tagHas) > 0)
}
