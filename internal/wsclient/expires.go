package wsclient

import (
	"errors"
	"fmt"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/api"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/db"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/settings"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/slack"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/utils"
)

func (wsc *wsclient) Authenticate() error {
	authToken := wsc.ws.GetURLQuery("auth_token")

	if authToken == "" {
		return errors.New("auth token is empty")
	}
	esUser := models.ESUser{}

	if err := api.NewAPI(settings.GetEBOAPI().String() + "/v1/validate-token/").
		SetIdentifier(string(wsc.identifier) + " Authenticate").
		AddHeaders(map[string]string{
			"User-Agent":    settings.GetUserAgent().String(),
			"Authorization": settings.GetServerToken().String(),
			"Content-Type":  "application/json",
		}).
		AddBody(map[string]string{
			"token": authToken,
		}).
		Post(&esUser); err != nil {
		if err.GetResponse() != nil && err.GetResponse().StatusCode >= 500 {
			go slack.SendPayload(slack.NewLootboxNotification(
				slack.IdentifierToTitle(string(wsc.identifier))+"api authenticate",
				fmt.Sprint("*Status Code:* \n> *", err.GetResponse().StatusCode, "*\n*Response Body:* \n> *", err.GetResponseBody(), "*"),
			), slack.LootboxHealthCheck)
		}
		return err
	}
	user := models.User{
		EsportsID:       esUser.ID,
		IsAccountFrozen: false,
	}

	if err := db.Shared().Where(user).First(&user).Error; err != nil {
		return err
	}

	if !user.IsActive {
		return errors.New("account is not active")
	}
	user.AuthToken = &authToken
	user.UserRequest = wsc.ws.GetUserRequest()
	wsc.user = &user
	wsc.ResetActivityExpiration() //added here since authenticate is called on connect and ws request
	return nil
}

// activity (60 000 milliseconds)
func (wsc *wsclient) ResetActivityExpiration() {
	wsc.activityExpirationTS = utils.TimeNow().UnixMilli() + 60_000 //added 60s TODO: add to constants
}

func (wsc *wsclient) IsActivityExpired() bool {
	return false
	// return utils.TimeNow().UnixMilli() > wsc.activityExpirationTS
}

// update table (60 000 milliseconds)
func (wsc *wsclient) UpdateTableIfNeeded() {
	if utils.TimeNow().UnixMilli() > wsc.updateTableExpirationTS {
		go func() {
			if err := api.NewAPI(settings.GetMGCoreAPI().String() + "/game-client/mini-game/v2/tables/").
				SetIdentifier(string(wsc.identifier) + " UpdateTable").
				AddHeaders(map[string]string{
					"User-Agent":    settings.GetUserAgent().String(),
					"Authorization": "Token " + *wsc.user.AuthToken,
					"Content-Type":  "application/json",
				}).
				Get(nil); err != nil {
				if err.GetResponse() != nil && err.GetResponse().StatusCode >= 500 {
					slack.SendPayload(slack.NewLootboxNotification(
						slack.IdentifierToTitle(string(wsc.identifier))+"api updateTable",
						fmt.Sprint("*Status Code:* \n> *", err.GetResponse().StatusCode, "*\n*Response Body:* \n> *", err.GetResponseBody(), "*"),
					), slack.LootboxHealthCheck)
				}
				return
			}
		}()
		wsc.updateTableExpirationTS = utils.TimeNow().UnixMilli() + 60_000
	}
}
