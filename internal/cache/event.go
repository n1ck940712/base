package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/constants"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
)

func GetEvent(tableID int64) (*models.Event, error) {
	key := fmt.Sprint(tableID, "-cevent")
	rEvent, err := Main.GetOrig(key)

	if err != nil {
		return nil, err
	}
	event := models.Event{}

	if err := json.Unmarshal([]byte(rEvent), &event); err != nil {
		return nil, err
	}

	if event.ID == nil {
		return nil, errors.New("event ID is nil, key: " + key + " save data not found")
	}

	return &event, nil
}

func SaveEvent(tableID int64, event *models.Event) error {
	key := fmt.Sprint(tableID, "-cevent")
	saveObj, err := json.Marshal(event)

	if err != nil {
		logger.Error("SaveEvent error: ", err.Error())
		return err
	}
	Main.Set(key, saveObj, time.Second*constants.LOL_TOWER_GAME_DURATION)
	return nil
}
