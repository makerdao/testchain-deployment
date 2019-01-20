package command

import (
	"encoding/base64"
	"encoding/json"
)

// Error is command error wrapper
type Error struct {
	Message error
	Stderr  []byte
}

func (e Error) Error() string {
	return e.Message.Error() + "\n" + string(e.Stderr)
}

//NewError init command error
func NewError(message error, stderr []byte) *Error {
	return &Error{Message: message, Stderr: stderr}
}

func (e Error) MarshalJSON() ([]byte, error) {
	m := struct {
		Message string `json:"message"`
		Stderr  string `json:"stderr"`
	}{}
	m.Message = e.Message.Error()
	m.Stderr = base64.StdEncoding.EncodeToString(e.Stderr)
	return json.Marshal(m)
}
