package cache

import (
	"encoding/json"
	"fmt"
	"time"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
)

func GetLeaderboard(eventID int64, prevEventID int64, leaderboard any) error {
	curKey := fmt.Sprint("leaderboard-", eventID)
	curSavedObj, err := Main.GetOrig(curKey)

	if err == nil {
		if err := json.Unmarshal([]byte(curSavedObj), leaderboard); err != nil {
			return err
		}
		return nil
	}

	prevKey := fmt.Sprint("leaderboard-", prevEventID)
	prevSavedObj, pErr := Main.GetOrig(prevKey)

	if pErr != nil {
		return pErr
	}
	if err := json.Unmarshal([]byte(prevSavedObj), leaderboard); err != nil {
		return err
	}
	return nil
}

func SaveLeaderboard(eventID int64, leaderboard any) error {
	key := fmt.Sprint("leaderboard-", eventID)
	saveObj, err := json.Marshal(leaderboard)

	if err != nil {
		logger.Error("SaveLeaderboard error: ", err.Error())
		return err
	}

	Main.Set(key, saveObj, time.Duration(2)*time.Minute)
	return nil
}
