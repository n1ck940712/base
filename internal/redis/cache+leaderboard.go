package redis

import (
	"encoding/json"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/models/response"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/utils"
	r "github.com/go-redis/redis/v8"
)

const (
	leaderboardBaseKey = "leaderboard_"
)

func SetLeaderboard(identfier string, leaderboard *response.Leaderboard) error {
	return Cache().Set(leaderboardBaseKey+identfier, utils.JSON(leaderboard), r.KeepTTL)
}

func GetLeaderboard(identfier string) (*response.Leaderboard, error) {
	if cLeaderboard, err := Cache().Get(leaderboardBaseKey + identfier); err != nil {
		return nil, err
	} else {
		leaderboard := (*response.Leaderboard)(nil)

		if err := json.Unmarshal([]byte(cLeaderboard), &leaderboard); err != nil {
			return leaderboard, err
		}
		return leaderboard, nil
	}
}
