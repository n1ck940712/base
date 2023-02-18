package gamemanager

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/cache"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/constants"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/errors"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/service"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/settings"
	utils "bitbucket.org/esportsph/minigame-backend-golang/internal/utils"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/hashutil"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/types"
)

var (
	iSettings        settings.ISettings                 = settings.NewSettings()
	iTicketService   service.TicketService              = *service.NewTicket()
	iUserService     service.UserService                = *service.NewUser()
	ILolTowerService service.LolTowerMemberLevelService = *service.NewLolTowerMemberLevelService()
	iEventService    service.EventService               = service.NewEvent()
	iMemberConfig    service.MemberConfigService        = service.NewMemberConfig()
)

type LolTowerGameManager struct {
	gameID          int64
	tableID         int64
	reqTicket       *ReqTicket
	ticketSelection *TicketSelection
	userMgDetails   *models.User
}

func NewLolTowerGameManager(tableID int64) *LolTowerGameManager {
	return &LolTowerGameManager{
		gameID:          constants.LOL_TOWER_GAME_ID,
		tableID:         tableID,
		ticketSelection: nil,
	}
}

func (l *LolTowerGameManager) CreateFutureEvents() {
	created := createFutureEvents(l.gameID, l.tableID, constants.MAX_FUTURE_EVENTS)

	if created {
		logger.Info("LOL TOWER Events Generated")
	}
}

func (l *LolTowerGameManager) GetCurrentEvent() *models.Event {
	return eventService.GetCurrentEventByTableID(l.tableID, l.gameID)
}

func (l *LolTowerGameManager) GetCurrentEventAllowedForBet() {

}

func (l *LolTowerGameManager) ProcessSelection(param map[string]interface{}) (interface{}, error) {
	return true, nil
}

func (l *LolTowerGameManager) HandleTicket(channel string, param map[string]interface{}, user *models.User) (*ReqTicket, *TicketSelection, errors.FinalErrorMessage) {
	value := param
	level := int16(constants.LOL_TOWER_DEFAULT_LEVEL)

	btype, _ := json.Marshal(user)
	json.Unmarshal(btype, &l.userMgDetails)

	if channel == "selection" {
		levelSkipCnt := iTicketService.GetMemberLevel(param)
		level = levelSkipCnt.Level
		logger.Info("levelSkip ----", levelSkipCnt)
		if level == 0 {
			logger.Info("!!Warning --------- Level is equal to zero but in combo")
		}

		data := param["data"].(map[string]interface{})
		if selection, ok := data["selection"].(string); ok {
			if err := l.setTicketSelection(selection, levelSkipCnt); err != nil {
				return nil, nil, err
			}
		} else { //selection is not string
			return nil, nil, errors.FinalizeErrorMessage(errors.VALIDATE_SELECTION_ERROR, errors.IEID_INVALID_SELECTION, false)
		}
	} else {
		tbyte, _ := json.Marshal(value["data"])
		json.Unmarshal(tbyte, &l.reqTicket)
		euroOdds := getOdds(level)
		l.reqTicket.Tickets[0].EuroOdds = euroOdds
		l.reqTicket.Tickets[0].HongkongOdds = *types.Odds(euroOdds).EuroToHK(2).Ptr()
		l.reqTicket.Tickets[0].ReferenceNo = uuid.NewString()
		// level++ // increment
		l.reqTicket.Tickets[0].Level = level
	}

	return l.reqTicket, l.ticketSelection, nil
}

func (l *LolTowerGameManager) setTicketSelection(selection string, val service.LevelSkipCnt) errors.FinalErrorMessage {
	event, eErr := cache.GetEvent(l.tableID)

	if eErr != nil {
		return errors.FinalizeErrorMessage(errors.EVENT_STATUS_ERROR, errors.IEID_EVENT_NOT_FOUND, false)
	}
	level := val.Level

	if utils.InArray(selection, constants.LOL_BET_SELECTION) || selection == constants.CUSTOM_WIN_SELECTION || selection == constants.CUSTOM_LOSS_SELECTION {
		level++
	}

	activeTicket := iTicketService.GetMemberActiveTicket(l.userMgDetails.ID)

	if activeTicket == nil {
		return errors.FinalizeErrorMessage(errors.VALIDATE_SELECTION_ERROR, errors.IEID_INVALID_SELECTION, false)
	}
	euroOdds := getOdds(level)

	activeTicket.Odds = *types.Odds(euroOdds).EuroToHK(2).Ptr()
	activeTicket.EuroOdds = euroOdds
	activeTicket.EventID = *event.ID
	activeTicket.Selection = selection

	l.ticketSelection = &TicketSelection{
		Event:        event,
		ActiveTicket: *activeTicket,
		ComboTicket:  *iTicketService.GetMemberLatestComboTicket(activeTicket.ID),
		Level:        level,
		Skip:         val.Skip,
		Selection:    selection,
	}
	return nil
}

func getOdds(key int16) float64 {
	return iSettings.Get().LOL_LEVELS[int(key)]
}

/*
Get
*/
func (l *LolTowerGameManager) GenerateTowerMemberLevel(cticket models.ComboTicket, user *models.User, skip bool, curLevel int16, skipCnt int8) models.LolTowerMemberLevel {
	if skip {
		skipCnt--
	}

	time := time.Now()

	data := models.LolTowerMemberLevel{
		TicketID:      cticket.TicketID,
		ComboTicketID: cticket.ID,
		UserID:        user.ID,
		Level:         int8(curLevel),
		Skip:          skipCnt,
		NextLevelOdds: settings.LOL_LEVELS[int(curLevel)+1],
		Ctime:         time,
		Mtime:         time,
	}

	ILolTowerService.CreateLOLTowerMemberLevel(data)
	return data
}

func generateHashResult(hash string, gameID int64) (Result string, SelectionHeaderID int64) {
	result, err := hashutil.LOLTowerGenerateResult(hashutil.NewHash(hash))

	if err != nil {
		logger.Error("loltower LOLTowerGenerateResult error: ", err.Error())
	}
	// get result selection line
	sl := selectionService.GetSelection(gameID)

	return result, sl.ID
}

func (l *LolTowerGameManager) GetMemberTicketState(userID int64) (*service.TicketState, error) {
	userDetails, _ := iUserService.GetMGDetails(userID)
	res, err := ILolTowerService.GetMemberTicketStateV3(l.tableID, userDetails.ID)

	return res, err
}

func (l *LolTowerGameManager) IsEventOpenForBet(eventID int64) bool {
	event := iEventService.GetLastEventByEventID(eventID)
	timeNow := time.Now()
	sTime := event.StartDatetime
	eTime := event.StartDatetime.Add(constants.LOL_TOWER_BETTING_DURATION * time.Second)

	return utils.InTimeSpan(sTime, eTime, timeNow)
}

func (l *LolTowerGameManager) GetOdds() map[string]interface{} {
	levelOdds := make(map[string]interface{})

	for lvl, euroOdds := range settings.LOL_LEVELS {
		levelOdds["level"+strconv.Itoa(lvl)] = types.Odds(euroOdds).EuroToHK(2).Ptr()
	}

	return levelOdds
}

func (l *LolTowerGameManager) SettleTickets(event models.Event) {
	event.Status = constants.EVENT_STATUS_SETTLEMENT_IN_PROGRESS
	eventService.UpdateEventStatus(*event.ID, event)

	l.settleLossTickets(*event.ID)
	l.settleUnattendedTickets(*event.ID, *event.PrevEventID)
	l.settleMaxLevelTickets(*event.ID)
	l.settleMaxPayoutTickets(*event.ID)
	l.settleOldUnsettledTickets(event)

	event.Status = constants.EVENT_STATUS_SETTLED
	eventService.UpdateEventStatus(*event.ID, event)

	// Proccess Daily Max Winnings
	l.ProcessDailyMaxWinnings()
}
func (l *LolTowerGameManager) SettleUnsettledPrevEvents(event models.Event) {
	iEventService.SettleUnsettledPrevEvents(event, l.tableID)
}

func (l *LolTowerGameManager) settleOldUnsettledTickets(event models.Event) {
	tickets := iTicketService.GetOldUnsettledTicketsForSettlement(event, l.tableID)
	var ticketIDs []string
	for _, ticket := range *tickets {
		ticketIDs = append(ticketIDs, ticket.ID)
	}
	if ticketIDs != nil {
		updateTickets(ticketIDs)
	}
}

func (l *LolTowerGameManager) settleMaxPayoutTickets(eventID int64) {
	tickets := iTicketService.GetMaxPayoutTickets(eventID, l.tableID)

	var ticketIDs []string
	for _, ticket := range *tickets {
		ticketIDs = append(ticketIDs, ticket.TicketID)
	}
	if ticketIDs != nil {
		updateTickets(ticketIDs)
	}
}

func (l *LolTowerGameManager) settleMaxLevelTickets(eventID int64) {
	tickets := iTicketService.GetEventMaxLevelTickets(eventID, l.tableID)

	var ticketIDs []string
	for _, ticket := range *tickets {
		ticketIDs = append(ticketIDs, ticket.TicketID)
	}
	if ticketIDs != nil {
		updateTickets(ticketIDs)
	}
}

func (l *LolTowerGameManager) settleLossTickets(eventID int64) {
	tickets := iTicketService.GetEventLossTickets(eventID)

	var ticketIDs []string
	for _, ticket := range *tickets {
		ticketIDs = append(ticketIDs, ticket.TicketID)
	}

	if ticketIDs != nil {
		updateTickets(ticketIDs)
	}
}

func (l *LolTowerGameManager) settleUnattendedTickets(eventID int64, prevEventID int64) {
	logger.Info("checked ignored/unattended ticket----------")
	tickets := iTicketService.GetUnattendedTickets(eventID, prevEventID, l.tableID)
	var ticketIDs []string
	if tickets != nil {
		for _, ticket := range *tickets {
			ticketIDs = append(ticketIDs, ticket.TicketID)
		}
	}

	if ticketIDs != nil {
		updateTickets(ticketIDs)
	}
}

func updateTickets(ticketIDs []string) {
	iTicketService.UpdateMultipleTickets(ticketIDs, constants.TICKET_STATUS_SETTLED_PENDING_PAYOUT)
	iTicketService.UpdateComboTickets(ticketIDs, constants.TICKET_STATUS_SETTLED_PENDING_PAYOUT)
}

type Champion struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func (l *LolTowerGameManager) GetLeaderboards() (string, []Champion) {
	//set 14 + 7 + 4 for leaderboards
	var leaderBoards = []models.LeaderBoard{}
	var champions = []Champion{}
	offset := constants.LOL_TOWER_GAME_DURATION + constants.LOL_TOWER_BETTING_DURATION + constants.LOL_TOWER_STOP_BETTING_DURATION
	offset2 := constants.LOL_TOWER_BETTING_DURATION + constants.LOL_TOWER_STOP_BETTING_DURATION

	leaderBoardLevels := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9"}
	leaderBoards, _ = ILolTowerService.GetLeaderboards(leaderBoardLevels, offset, offset2, l.tableID)
	level := ""

	for _, leaderBoard := range leaderBoards {
		level = leaderBoard.Level
		champions = append(champions, Champion{
			Id:   leaderBoard.GetEncryptedUserID(),
			Name: leaderBoard.GetEncryptedName(),
		})
	}

	return level, champions
}

func (l *LolTowerGameManager) GetChampionMembers() ([]Champion, error) {
	//set 14 for champions
	var leaderBoards = []models.LeaderBoard{}
	var champions = []Champion{}
	offset2 := constants.LOL_TOWER_BETTING_DURATION + constants.LOL_TOWER_STOP_BETTING_DURATION

	leaderBoards, err := ILolTowerService.GetLeaderboards([]string{"10"}, constants.LOL_TOWER_GAME_DURATION, offset2, l.tableID)

	for _, leaderBoard := range leaderBoards {
		champions = append(champions, Champion{
			Id:   leaderBoard.GetEncryptedUserID(),
			Name: leaderBoard.GetEncryptedName(),
		})
	}

	return champions, err
}

func (l *LolTowerGameManager) IsGreaterThanMaxPayout(userID int64) bool {
	details, err := iTicketService.GetMemberTicketPayoutDetails(userID, l.tableID, constants.LOL_TOWER_GAME_DURATION+constants.LOL_TOWER_BETTING_DURATION)
	if err != nil {
		return true
	}
	euroOdds := settings.LOL_LEVELS[int(details.CurrentLevel)]

	return utils.CalculateMaxPayout(details.BetAmount, euroOdds) > details.MaxPayout
}

func (l *LolTowerGameManager) GetResult(selection string, eventID int64) interface{} {
	var res string

	if selection == "w" {
		s := iTicketService.GetEventResult(eventID)
		res = s[0]
	} else if selection == "l" {
		bomb := iEventService.GetEventLolTowerBomb(eventID)
		s := strings.Split(bomb, ",")
		res = s[0]
	}

	return res
}

func (l *LolTowerGameManager) ValidateSelected(data interface{}, isBetMessage bool) errors.FinalErrorMessage {
	dType := fmt.Sprintf("%T", data)

	if dType == "string" {
		if isBetMessage {
			if !utils.InArray(data.(string), constants.LOL_BET_SELECTION) {
				return errors.FinalizeErrorMessage(errors.VALIDATE_SELECTION_ERROR, errors.IEID_INVALID_SELECTION, false)
			}
		} else {
			allSelection := append(constants.LOL_BET_SELECTION, constants.LOL_EXTRA_SELECTION...)
			if !utils.InArray(data.(string), allSelection) {
				return errors.FinalizeErrorMessage(errors.VALIDATE_SELECTION_ERROR, errors.IEID_INVALID_SELECTION, false)
			}
		}
	} else {
		return errors.FinalizeErrorMessage(errors.VALIDATE_SELECTION_ERROR, errors.IEID_INVALID_SELECTION, false)
	}

	return nil
}

func (l *LolTowerGameManager) GetTableConfig(userID int64) (Config, errors.FinalErrorMessage) {
	event, eErr := cache.GetEvent(l.tableID)

	if eErr != nil {
		return Config{}, errors.FinalizeErrorMessage(errors.EVENT_STATUS_ERROR, errors.IEID_EVENT_NOT_FOUND, false)
	}

	// chipSet := settings.LOL_CHIPSET
	user, mgDetailsError := iUserService.GetMGDetails(userID)
	betChipset := models.BetChipset{
		Currency:      user.CurrencyCode,
		CurrencyRatio: user.CurrencyRatio,
		TableID:       l.tableID,
	}
	if err := service.Get(&betChipset); err != nil {
		betChipset.CurrencyRatio = 0
		if err := service.Get(&betChipset); err != nil {
			betChipset.TableID = 0
			betChipset.Default = true
			if err := service.Get(&betChipset); err != nil {
				betChipset.Default = false
				if err := service.Get(&betChipset); err != nil {
					logger.Error("bet chipset error: ", err.Error())
				}
			}
		}
	}
	gt := memberTableService.GetMemberTable(user.ID, l.tableID)
	curEvent := CurrentEvent{}

	if event == nil {
		return Config{}, errors.FinalizeErrorMessage(errors.VALIDATE_EVENT_ERROR, errors.IEID_EVENT_NOT_FOUND, false)
	} else {
		curEvent = CurrentEvent{
			ID:            *event.ID,
			StartDatetime: event.StartDatetime,
		}
	}
	conf := Config{
		ID:              gt.ID,
		MaxBetAmount:    gt.MaxBetAmount,
		MaxPayoutAmount: gt.MaxPayoutAmount,
		MinBetAmount:    gt.MinBetAmount,
		BetChips:        betChipset.GetBetChips(),
		CurrentEvent:    curEvent,
		Enable:          gt.IsEnabled,
		IsAnonymous:     false,
		EnableAutoPlay:  gt.IsAutoPlayEnabled,
		ResultAnimation: true,
		EffectsSound:    0.5,
		GameSound:       0.5,
	}
	//bugfix : getting different config if user doesn't exists.
	if mgDetailsError != nil {
		return conf, nil
	}

	conf.setMemberConfig(l.gameID, user.ID)
	return conf, nil
}

func (l *LolTowerGameManager) GetSelections() []string {
	return constants.LOL_BET_SELECTION
}
