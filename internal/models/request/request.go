package request

import (
	"encoding/json"
)

type RequestData interface {
}

type Request struct {
	Type        string       `json:"type"`
	Data        *RequestData `json:"data"`
	RawJSONData []byte       //stored raw json of data
}

func NewRequest() *Request {
	return &Request{}
}

func (r *Request) ParseJSON(jsonStr string) error {
	if err := json.Unmarshal([]byte(jsonStr), r); err != nil {
		return err
	}
	if r.Data != nil { //store raw json data if not nil
		if rawJSON, err := json.Marshal(r.Data); err == nil {
			r.RawJSONData = rawJSON
		}
	}

	return nil
}
