package github

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

const basePath = "https://api.github.com"
const releaseURITemplate = "/repos/makerdao/%s/releases/tags/%s"
const resTagURITemplate = "/repos/makerdao/%s/git/refs/tags/%s"
const downloadTarGzSourceCode = "https://github.com/makerdao/%s/archive/%s.tar.gz"
const workFileName = "deploy.tar.gz"

//Client of github.com
type Client struct {
	cfg    Config
	client *http.Client
}

//NewClient init client
func NewClient(cfg Config) *Client {
	return &Client{
		cfg:    cfg,
		client: &http.Client{Timeout: time.Duration(cfg.ClientTimeoutInSecond) * time.Second},
	}
}

//GetDirInArchiveFromCfg return dir in archive downloaded from github
func (c *Client) GetDirInArchiveFromCfg() string {
	return fmt.Sprintf("%s-%s", c.cfg.RepoName, c.cfg.TagName)
}

//RemoveArchiveIfExists skip error if file not exists
func (c *Client) RemoveArchiveIfExists() error {
	path := filepath.Join(c.cfg.WorkDir, workFileName)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(path)
}

// DownloadTarGzSourceCode -=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-

// DownloadTarGzSourceCode download code source
func (c *Client) DownloadTarGzSourceCode(log *logrus.Entry) (string, error) {
	filePath := filepath.Join(c.cfg.WorkDir, workFileName)
	err := os.Remove(filePath)
	if !os.IsNotExist(err) {
		return "", err
	}
	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.WithError(err).Error("Can't close output file")
		}
	}()

	urlStr := fmt.Sprintf(downloadTarGzSourceCode, c.cfg.RepoName, c.cfg.TagName)
	httpReq, err := http.NewRequest(http.MethodGet, urlStr, nil)
	if err != nil {
		return "", err
	}
	httpReq.Header.Add("Cache-Control", "no-cache")
	httpReq.Header.Add("Authorization", "token "+c.cfg.APIToken)
	httpResp, err := c.client.Do(httpReq)
	if err != nil {
		return "", err
	}

	defer func() {
		if err := httpResp.Body.Close(); err != nil {
			log.WithError(err).Error("Can't close resp body")
		}
	}()

	_, err = io.Copy(file, httpResp.Body)
	if err != nil {
		return "", err
	}

	return filePath, nil
}

// GetTagDownloadURL -=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-

// GetTagResponse is response data
type GetTagResponse struct {
	Assets []GetTagResponseAsset `json:"assets"`
}

// GetTagResponseAsset is part with asset of response data
type GetTagResponseAsset struct {
	BrowserDownloadURL string `json:"browser_download_url"`
}

// GetTagDownloadURL return url for download repo or err if has any problem
func (c *Client) GetTagDownloadURL(log *logrus.Entry) (string, error) {
	urlStr := fmt.Sprintf(basePath+releaseURITemplate, c.cfg.RepoName, c.cfg.TagName)
	httpReq, err := http.NewRequest(http.MethodGet, urlStr, nil)
	if err != nil {
		return "", err
	}
	httpReq.Header.Add("Cache-Control", "no-cache")
	httpReq.Header.Add("Authorization", "token "+c.cfg.APIToken)

	httpResp, err := c.client.Do(httpReq)
	if err != nil {
		return "", err
	}

	if httpResp.StatusCode != http.StatusOK {
		log.WithField("http_status_code", httpResp.StatusCode).Debug("unexpected http status code")
		return "", errors.New("unexpected http status code, expected OK")
	}

	respBodyBytes, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return "", err
	}
	if err := httpResp.Body.Close(); err != nil {
		log.WithError(err).Error("Can't close body of response")
	}

	var resp GetTagResponse
	if err := json.Unmarshal(respBodyBytes, &resp); err != nil {
		return "", err
	}

	if len(resp.Assets) != 1 {
		log.WithField("len_of_assets", len(resp.Assets)).Debug("Unexpected len of assets in response")
		return "", errors.New("unexpected len of assets in response")
	}

	return resp.Assets[0].BrowserDownloadURL, nil
}

// GetTagHash -=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=

// GetTagHashResponse is response data
type GetTagHashResponse struct {
	Object struct {
		Sha string `json:"sha"`
	} `json:"object"`
}

// GetTagHash return hash for configured tag
func (c *Client) GetTagHash(log *logrus.Entry) (string, error) {
	log = log.WithField("action", "GetTagHash")
	urlStr := fmt.Sprintf(basePath+resTagURITemplate, c.cfg.RepoName, c.cfg.TagName)
	httpReq, err := http.NewRequest(http.MethodGet, urlStr, nil)
	if err != nil {
		return "", err
	}
	httpReq.Header.Add("Cache-Control", "no-cache")
	httpReq.Header.Add("Authorization", "token "+c.cfg.APIToken)

	httpResp, err := c.client.Do(httpReq)
	if err != nil {
		return "", err
	}

	if httpResp.StatusCode != http.StatusOK {
		log.WithField("http_status_code", httpResp.StatusCode).Debug("unexpected http status code")
		return "", errors.New("unexpected http status code, expected OK")
	}

	respBodyBytes, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return "", err
	}
	if err := httpResp.Body.Close(); err != nil {
		log.WithError(err).Error("Can't close body of response")
	}

	var resp GetTagHashResponse
	if err := json.Unmarshal(respBodyBytes, &resp); err != nil {
		return "", err
	}

	return resp.Object.Sha, nil
}
