package placebet

import (
	"bitbucket.org/esportsph/minigame-backend-golang/internal/errors"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/gamemanager"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
)

type IPlacebet interface {
	Lock() bool
	Unlock()
	ProcessTicket(value map[string]interface{}) (*models.ResponseData, errors.FinalErrorMessage)
	ProcessSelection(value map[string]interface{}) (interface{}, errors.FinalErrorMessage)
	ValidateRequest(ignoreHasticket bool, validateBalance bool, gameManager gamemanager.IGameManager) errors.FinalErrorMessage
}
