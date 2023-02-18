package api

import (
	"encoding/json"

	constants_lolcouple "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/lolcouple"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/settings"
)

var loadDummyResponse = false //useful for returning response without internet

//used for loading responses;
//only used if settings.ENVIRONMENT is "local" and loadDummyResponse is true and has dummyResponses() object
func (a *api) dummyExecute(response any) bool {
	if settings.ENVIRONMENT == "local" && loadDummyResponse {
		if jsonString, ok := dummyResponses()[a.identifier]; ok {

			json.Unmarshal([]byte(jsonString), &response)

			return true
		}
	}
	return false
}

//map[string:{identifier}]string:{jsonString}
func dummyResponses() map[string]string {
	return map[string]string{
		constants_lolcouple.Identifier + " authenticate": `{"type":"member","id":114258,"metadata":{"member_code":"testjus","partner_id":2,"is_account_frozen":false,"currency_code":"RMB","exchange_rate":1.0,"currency_ratio":1.0}}`,
	}
}
