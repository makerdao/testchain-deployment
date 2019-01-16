package protocol

import (
	"encoding/json"

	"github.com/makerdao/testchain-deployment/pkg/serror"
)

//Request wrapper for rpc
type Request struct {
	ID     string          `json:"id"`
	Method string          `json:"method"`
	Data   json.RawMessage `json:"data"`
}

//ResponseType is type for rpc response
type ResponseType string

const (
	ResponseTypeOK  = "ok"
	ResponseTypeErr = "error"
)

//Response wrapper for rpc
type Response struct {
	Type   ResponseType    `json:"type"`
	Result json.RawMessage `json:"result"`
}

//GetMarshalResponseErrorBytes static error if we have problem with protocol
func GetMarshalResponseErrorBytes() []byte {
	return []byte(`{
	"type": "` + ResponseTypeErr + `"
	"result": {
		"code": "` + serror.ErrCodeInternalError + `",
		"detail": "Can't marshal response body'",
		"errorList": []
	}
}`)
}
