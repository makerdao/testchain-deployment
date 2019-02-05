package deploy

import (
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/makerdao/testchain-deployment/pkg/command"
)

//StepModel - we put data from json to that struct
type StepModel struct {
	ID               int             `json:"id,omitempty"`
	Description      string          `json:"description"`
	OmniaFromAddress string          `json:"omniaFromAddr"`
	Defaults         json.RawMessage `json:"defaults"`
	Roles            json.RawMessage `json:"roles"`
	Oracles          json.RawMessage `json:"oracles"`
	Ilks             json.RawMessage `json:"ilks"`
}

type ResultErrorModel struct {
	Msg       string `json:"msg"`
	StderrB64 string `json:"stderrB64"`
}

func NewResultErrorModelFromErr(err error) *ResultErrorModel {
	return &ResultErrorModel{
		Msg: err.Error(),
	}
}

func NewResultErrorModelFromTxt(msg string) *ResultErrorModel {
	return &ResultErrorModel{
		Msg: msg,
	}
}

func NewResultErrorModelFromCmd(err *command.Error) *ResultErrorModel {
	return NewResultErrorModelFromErr(err.Message).WithStderr(err.Stderr)
}

func (m *ResultErrorModel) WithStderr(stderr []byte) *ResultErrorModel {
	m.StderrB64 = base64.StdEncoding.EncodeToString(stderr)
	return m
}

//ResultModel is struct for result of run
type ResultModel struct {
	LastUpdated time.Time       `json:"lastUpdated"`
	Data        json.RawMessage `json:"data"`
}

//NewResultModel init model of result
func NewResultModel(lastUpdated time.Time, data json.RawMessage) *ResultModel {
	return &ResultModel{LastUpdated: lastUpdated, Data: data}
}
