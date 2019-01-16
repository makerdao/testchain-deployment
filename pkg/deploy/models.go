package deploy

import "encoding/json"

//Model - we put data from json to that struct
type Model struct {
	ID          int             `json:"id,omitempty"`
	Description string          `json:"description"`
	Defaults    json.RawMessage `json:"defaults"`
	Roles       json.RawMessage `json:"roles"`
	Oracles     json.RawMessage `json:"oracles"`
}
