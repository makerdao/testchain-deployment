package gateway

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/makerdao/testchain-deployment/pkg/service/nats"
	"github.com/makerdao/testchain-deployment/pkg/service/protocol"
	gonats "github.com/nats-io/go-nats"
	"github.com/sirupsen/logrus"
)

const (
	rpcURI = "/rpc"
)

// Client is http client for gateway service
type Client struct {
	basePath string
	client   *http.Client
	nats     *gonats.Conn
	natsCfg  nats.Config
}

// NewClient init client
func NewClient(cfg Config, nats *gonats.Conn, natsCfg nats.Config) *Client {
	return &Client{
		basePath: fmt.Sprintf("http://%s:%d", cfg.Host, cfg.Port),
		client:   &http.Client{Timeout: time.Duration(cfg.ClientTimeoutInSecond) * time.Second},
		nats:     nats,
		natsCfg:  natsCfg,
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
	return c.reqRegisterUnregister(
		log,
		"RegisterDeployment",
		req,
	)
}

// Unregister -=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=

// Unregister instance on gateway
func (c *Client) Unregister(log *logrus.Entry, req *ServiceData) error {
	return c.reqRegisterUnregister(
		log,
		"UnregisterDeployment",
		req,
	)
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

	_, httpErr := c.req(log, "RunResult", reqBytes)

	natsErr := c.nats.Publish(c.getPublishTopic("RunResult", req.ID), reqBytes)
	if natsErr != nil || httpErr != nil {
		return fmt.Errorf("http: %+v, nats: %+v", httpErr, natsErr)
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
	r.Type = UpdateResultRequestTypeErr
	r.Result = json.RawMessage(fmt.Sprintf(`{"message":"%s"}`, err.Error()))
	return r
}

// UpdateResult send result of update source
func (c *Client) UpdateResult(log *logrus.Entry, req *UpdateResultRequest) error {
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return err
	}

	_, httpErr := c.req(log, "UpdateResult", reqBytes)

	natsErr := c.nats.Publish(c.getPublishTopic("UpdateResult", req.ID), reqBytes)
	if natsErr != nil || httpErr != nil {
		return fmt.Errorf("http: %+v, nats: %+v", httpErr, natsErr)
	}

	return nil
}

// CheckoutResult -=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=

type CheckoutResultRequestType string

const (
	CheckoutResultRequestTypeOK  = "ok"
	CheckoutResultRequestTypeErr = "error"
)

type CheckoutResultRequest struct {
	ID     string               `json:"id"`
	Type   RunResultRequestType `json:"type"`
	Result json.RawMessage      `json:"result"`
}

func (r *CheckoutResultRequest) SetErr(err error) *CheckoutResultRequest {
	r.Type = RunResultRequestTypeErr
	r.Result = json.RawMessage(fmt.Sprintf(`{"message":"%s"}`, err.Error()))
	return r
}

// CheckoutResult send result of checkout to commit
func (c *Client) CheckoutResult(log *logrus.Entry, req *CheckoutResultRequest) error {
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return err
	}

	_, httpErr := c.req(log, "CheckoutResult", reqBytes)

	natsErr := c.nats.Publish(c.getPublishTopic("CheckoutResult", req.ID), reqBytes)
	if natsErr != nil || httpErr != nil {
		return fmt.Errorf("http: %+v, nats: %+v", httpErr, natsErr)
	}

	return nil
}

// -=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-

func (c *Client) reqRegisterUnregister(log *logrus.Entry, method string, req *ServiceData) error {
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
	log.WithField("Client.Method", "Gateway."+method)
	reqBody := protocol.Request{
		Method: method,
		Data:   reqBytes,
	}
	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	log.Debugf("Request data: %s", string(reqBodyBytes))
	httpReq, err := http.NewRequest(http.MethodPost, c.basePath+rpcURI, bytes.NewReader(reqBodyBytes))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Add("Content-Type", "application/json")

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

func (c *Client) getPublishTopic(name string, id string) string {
	return fmt.Sprintf("%s.%s.%s", c.natsCfg.TopicPrefix, name, id)
}
