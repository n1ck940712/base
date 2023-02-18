package request

import (
	"encoding/json"

	"bitbucket.org/esportsph/minigame-backend-golang/pkg/types"
)

type SelectionData struct {
	Selection types.JSONRaw `json:"selection"`
}

// json mirror of selection data
type JSONSelectionData struct {
	Selection json.RawMessage `json:"selection"`
}

func (jsd *JSONSelectionData) GetSelectionData() *SelectionData {
	return &SelectionData{
		Selection: types.JSONRaw(jsd.Selection),
	}
}

func (r *Request) GetSelectionData() *SelectionData {
	jsonSelectionData := JSONSelectionData{}

	json.Unmarshal(r.RawJSONData, &jsonSelectionData)
	return jsonSelectionData.GetSelectionData()
}
