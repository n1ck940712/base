package gamemanager

import (
	"bitbucket.org/esportsph/minigame-backend-golang/internal/errors"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/service"
)

type IGameManager interface {
	CreateFutureEvents()
	GetCurrentEvent() *models.Event
	GetCurrentEventAllowedForBet() // models.Event
	ProcessSelection(param map[string]interface{}) (interface{}, error)
	HandleTicket(channel string, val map[string]interface{}, user *models.User) (*ReqTicket, *TicketSelection, errors.FinalErrorMessage)
	GetMemberTicketState(userID int64) (*service.TicketState, error)
	GenerateTowerMemberLevel(combo models.ComboTicket, user *models.User, skip bool, level int16, skipCnt int8) models.LolTowerMemberLevel
	IsEventOpenForBet(eventID int64) bool
	GetOdds() map[string]interface{}
	SettleTickets(event models.Event)
	GetLeaderboards() (string, []Champion)
	GetChampionMembers() ([]Champion, error)
	IsGreaterThanMaxPayout(userId int64) bool
	GetResult(selection string, eventID int64) interface{}
	ValidateSelected(data interface{}, isBetMessage bool) errors.FinalErrorMessage
	GetTableConfig(userID int64) (Config, errors.FinalErrorMessage)
	GetSelections() []string
	UpsertConfig(userID int64, payload models.ConfigPatchRequestBody) (models.ConfigPatchRequestBody, error)
	ProcessDailyMaxWinnings()
	settleOldUnsettledTickets(event models.Event)
	SettleUnsettledPrevEvents(event models.Event)
}
