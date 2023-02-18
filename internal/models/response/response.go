package response

import (
	"encoding/json"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
)

type ResponseData interface {
	Description() string
}

type Response struct {
	Type        string       `json:"type,omitempty"`
	Data        ResponseData `json:"data,omitempty"`
	Cts         string       `json:"cts"`
	UserID      string       `json:"user_id,omitempty"`
	IsAnonymous *bool        `json:"is_anonymous,omitempty"`
	MemberCode  string       `json:"member_code,omitempty"`
	QueryParams ResponseData `json:"query_params,omitempty"`
}

func (r *Response) JSON() string {
	r.Cts = "cts_ts"
	if result, err := json.Marshal(r); err != nil {
		logger.Error("process Response json marshal error: ", err.Error())
		return ""
	} else {
		return string(result)
	}
}

type GenericMap map[string]any

func (GenericMap) Description() string {
	return "this is generic map!!"
}
