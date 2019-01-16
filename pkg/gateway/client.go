package gateway

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/makerdao/testchain-deployment/pkg/service/protocol"
	"github.com/sirupsen/logrus"
)

const (
	rpcURI = "/rpc"
)

// Client is http client for gateway service
type Client struct {
	basePath string
	client   *http.Client
}

// NewClient init client
func NewClient(cfg Config) *Client {
	return &Client{
		basePath: fmt.Sprintf("http://%s:%d", cfg.Host, cfg.Port),
		client:   &http.Client{Timeout: time.Duration(cfg.ClientTimeoutInSecond) * time.Second},
	}
}

// ServiceData is request data for register and unregister instance on gateway
type ServiceData struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

// Register -=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=

// Register instance on gateway
func (c *Client) Register(log *logrus.Entry, req *ServiceData) error {
	return c.reqREgisterUnregister(log, "RegisterDeployment", req)
}

// Unregister -=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=

// Unregister instance on gateway
func (c *Client) Unregister(log *logrus.Entry, req *ServiceData) error {
	return c.reqREgisterUnregister(log, "UnregisterDeployment", req)
}

// RunResult -=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-

type RunResultRequestType string

const (
	RunResultRequestTypeOK  = "ok"
	RunResultRequestTypeErr = "error"
)

type RunResultRequest struct {
	ID     string               `json:"id"`
	Type   RunResultRequestType `json:"type"`
	Result json.RawMessage      `json:"result"`
}

func (r *RunResultRequest) SetErr(err error) *RunResultRequest {
	r.Type = RunResultRequestTypeErr
	r.Result = json.RawMessage(fmt.Sprintf(`{"message":"%s"}`, err.Error()))
	return r
}

// RunResult send result of run to gateway
func (c *Client) RunResult(log *logrus.Entry, req *RunResultRequest) error {
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return err
	}

	if _, err := c.req(log, "RunResult", reqBytes); err != nil {
		return err
	}

	return nil
}

// UpdateResult -=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=

type UpdateResultRequestType string

const (
	UpdateResultRequestTypeOK  = "ok"
	UpdateResultRequestTypeErr = "error"
)

type UpdateResultRequest struct {
	ID     string                  `json:"id"`
	Type   UpdateResultRequestType `json:"type"`
	Result json.RawMessage         `json:"result"`
}

func (r *UpdateResultRequest) SetErr(err error) *UpdateResultRequest {
	r.Type = RunResultRequestTypeErr
	r.Result = json.RawMessage(fmt.Sprintf(`{"message":"%s"}`, err.Error()))
	return r
}

// UpdateResult send result of update source
func (c *Client) UpdateResult(log *logrus.Entry, req *UpdateResultRequest) error {
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return err
	}

	if _, err := c.req(log, "UpdateResult", reqBytes); err != nil {
		return err
	}

	return nil
}

// -=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-

func (c *Client) reqREgisterUnregister(log *logrus.Entry, method string, req *ServiceData) error {
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return err
	}

	if _, err := c.req(log, method, reqBytes); err != nil {
		return err
	}

	return nil
}

func (c *Client) req(log *logrus.Entry, method string, reqBytes json.RawMessage) (json.RawMessage, error) {
	reqBody := protocol.Request{
		Method: method,
		Data:   reqBytes,
	}
	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest(http.MethodPost, c.basePath+rpcURI, bytes.NewReader(reqBodyBytes))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Add("ContentType", "application/json")

	httpResp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}

	if httpResp.StatusCode != http.StatusOK {
		log.WithField("http_status_code", httpResp.StatusCode).Debug("unexpected http status code")
		return nil, errors.New("unexpected http status code, expected OK")
	}

	respBytes, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}
	if err := httpResp.Body.Close(); err != nil {
		log.WithError(err).Error("Can't close request body")
	}

	var respBody protocol.Response
	if err := json.Unmarshal(respBytes, &respBody); err != nil {
		return nil, err
	}
	if respBody.Type != protocol.ResponseTypeOK {
		log.Errorf("Gateway response error, %s", string(respBody.Result))
		return nil, errors.New("error from gateway")
	}

	return respBody.Result, nil
}
