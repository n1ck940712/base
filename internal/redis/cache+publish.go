package redis

import (
	"encoding/json"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/models/response"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/utils"
	r "github.com/go-redis/redis/v8"
)

const (
	publishStateBaseKey = "publish_state_"
)

func SetPublishState(identfier string, state *response.State) error {
	return Cache().Set(publishStateBaseKey+identfier, utils.JSON(state), r.KeepTTL)
}

func GetPublishState(identfier string) (*response.State, error) {
	if cState, err := Cache().Get(publishStateBaseKey + identfier); err != nil {
		return nil, err
	} else {
		state := (*response.State)(nil)

		if err := json.Unmarshal([]byte(cState), &state); err != nil {
			return state, err
		}
		return state, nil
	}
}
