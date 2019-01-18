package deploy

import (
	"encoding/base64"
	"encoding/json"
)

//StepModel - we put data from json to that struct
type StepModel struct {
	ID          int             `json:"id,omitempty"`
	Description string          `json:"description"`
	Defaults    json.RawMessage `json:"defaults"`
	Roles       json.RawMessage `json:"roles"`
	Oracles     json.RawMessage `json:"oracles"`
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

func (m *ResultErrorModel) WithStderr(stderr []byte) *ResultErrorModel {
	m.StderrB64 = base64.StdEncoding.EncodeToString(stderr)
	return m
}
