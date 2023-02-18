package process_bet

import (
	"encoding/json"
	"math/rand"
	"strings"
	"time"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/constants"
	constants_loltower "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/loltower"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/errors"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models/request"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models/response"
	process_bet "bitbucket.org/esportsph/minigame-backend-golang/internal/process/process-bet"
	process_wallet "bitbucket.org/esportsph/minigame-backend-golang/internal/process/process-wallet"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/settings"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/utils"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/types"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BetDatasouce interface {
	process_bet.BetDatasouce
}

type BetProcess interface {
	process_bet.BetProcess
	PlaceSelection(selectionData *request.SelectionData) response.ResponseData
}

type betProcess struct {
	datasource        BetDatasouce
	processDatasource process_bet.BetProcessDatasource
	walletProcess     process_wallet.WalletProcess
	prevMemberLevel   *models.LolTowerMemberLevel
}

func NewBetProcess(datasource BetDatasouce) BetProcess {
	betProcess := betProcess{
		datasource:        datasource,
		processDatasource: process_bet.NewBetProcessDatasource(datasource),
	}

	betProcess.walletProcess = process_wallet.NewWalletProcess(&betProcess)
	betProcess.processDatasource.SetBetOpenRangeCallback(betProcess.BetOpenRangeCallback)
	betProcess.processDatasource.SetBetResultCallback(betProcess.BetResultCallback)
	return &betProcess
}

func (bp *betProcess) Placebet(betData *request.BetData) response.ResponseData {
	defer bp.CleanUp()
	if err := bp.processDatasource.PrePrepareValidation(); err != nil {
		return err
	}
	bp.processDatasource.LoadActiveTickets()
	if err := bp.OverrideBetDataIfNeeded(betData); err != nil {
		return response.ErrorIE(errors.VALIDATE_SELECTION_ERROR, errors.IEID_INVALID_SELECTION, process_bet.BetType)
	}
	if tickets := bp.processDatasource.GetTickets(); tickets != nil {
		return response.ErrorIE(errors.DUPLICATE_BET, errors.IEID_DUPLICATE_BET, process_bet.BetType)
	}
	if err := bp.processDatasource.PrepareBet(betData); err != nil {
		return err
	}
	if err := bp.processDatasource.PreCreateTicketValidation(); err != nil {
		return err
	}
	if err := bp.processDatasource.PrepareTickets(); err != nil {
		return err
	}
	if err := bp.processDatasource.PrepareComboTickets(); err != nil {
		return err
	}
	if err := bp.walletProcess.CreateTransactions(); err != nil {
		return err
	}
	if err := bp.walletProcess.CommitTransactions(); err != nil {
		if err := bp.walletProcess.RollbackTransactions(); err != nil {
			return err
		}
		return err
	}
	if err := bp.processDatasource.CreateTicketsWithCombo(func(tx *gorm.DB) error {
		return bp.CreateTowerMemberLevel(tx)
	}); err != nil {
		if err := bp.walletProcess.RollbackTransactions(); err != nil {
			return err
		}
		return err
	}
	go bp.processDatasource.CheckDailyWinLoss()
	return &response.BetData{
		UserID:  bp.processDatasource.GetEncryptedUserID(),
		Tickets: bp.processDatasource.GetBetComboTickets(),
	}
}

func (bp *betProcess) PlaceSelection(selectionData *request.SelectionData) response.ResponseData {
	defer bp.CleanUp()
	if err := bp.processDatasource.PrePrepareValidation(); err != nil {
		return err
	}
	bp.processDatasource.SetProcessType(process_bet.SelectionType)
	bp.processDatasource.LoadActiveTickets()
	if err := bp.PrepareSelection(selectionData); err != nil {
		return err
	}
	if err := bp.processDatasource.PreCreateComboTicketValidation(); err != nil {
		return err
	}
	if resp := bp.ProcessPayoutSelectionIfNeeded(selectionData); resp != nil {
		return resp
	}
	if err := bp.processDatasource.PrepareComboTickets(); err != nil {
		return err
	}
	if err := bp.processDatasource.CreateComboTickets(func(tx *gorm.DB) error {
		tickets := bp.processDatasource.GetTickets()
		comboTickets := bp.processDatasource.GetComboTickets()
		prevMemberLevel := bp.GetPrevMemberLevel()
		totalEuroOdds := OddsFromLevel(utils.IfElse(selectionData.Selection.String() == constants_loltower.SelectionSkip, prevMemberLevel.Level, prevMemberLevel.Level+1))
		totalHongkongOdds := *totalEuroOdds.EuroToHK(2).Ptr()
		winLossAmount := -(*tickets)[0].Amount
		payoutAmount := 0.0

		if *(*comboTickets)[0].Result == constants.TICKET_RESULT_WIN {
			winLossAmount = totalHongkongOdds * (*comboTickets)[0].Amount
			payoutAmount = (*comboTickets)[0].Amount + winLossAmount
		}
		if err := tx.Updates(models.Ticket{
			ID:            (*tickets)[0].ID,
			EventID:       (*comboTickets)[0].EventID,
			Odds:          totalHongkongOdds,
			EuroOdds:      *totalEuroOdds.Precision(2).Ptr(),
			OriginalOdds:  &totalHongkongOdds,
			Result:        (*comboTickets)[0].Result,
			WinLossAmount: &winLossAmount,
			PayoutAmount:  &payoutAmount,
			Selection:     (*comboTickets)[0].Selection,
		}).Error; err != nil {
			return err
		}
		return bp.CreateTowerMemberLevel(tx)
	}); err != nil {
		return err
	}
	go bp.processDatasource.CheckDailyWinLoss()
	tickets := bp.processDatasource.GetBetComboTickets()

	if tickets != nil && selectionData.Selection.String() == constants_loltower.SelectionSkip {
		prevMemberLevel := bp.GetPrevMemberLevel()

		// (&(*tickets)[0]).Level = &prevMemberLevel.Level
		// (&(*tickets)[0]).Skip = utils.Ptr(prevMemberLevel.Skip - 1)
		// (&(*tickets)[0]).NextLevelOdds = utils.Ptr(types.Odds(prevMemberLevel.NextLevelOdds).String())
		return &SkipSelectionResponse{
			UserID:        bp.processDatasource.GetEncryptedUserID(),
			TicketID:      (*tickets)[0].ID,
			ComboTicketID: (*tickets)[0].ComboTicketID,
			Level:         prevMemberLevel.Level + 1,
			Skip:          prevMemberLevel.Skip - 1,
			NextLevelOdds: prevMemberLevel.NextLevelOdds,
			Selection:     constants_loltower.SelectionSkip,
		}
	}
	return &response.BetData{
		UserID:  bp.processDatasource.GetEncryptedUserID(),
		Tickets: tickets,
	}
}

// internal function
func (bp *betProcess) PrepareSelection(selectionData *request.SelectionData) *response.ErrorData {
	eventID, _ := json.Marshal(bp.datasource.GetEvent().ID)
	tableID, _ := json.Marshal(bp.datasource.GetGameTable().ID)
	tickets := bp.processDatasource.GetTickets()

	if tickets == nil {
		return response.ErrorIE(errors.VALIDATE_BET_SELECTION_TYPE_ERROR, errors.IEID_NO_ACTIVE_TICKET, process_bet.SelectionType)
	}
	amount, _ := json.Marshal((*tickets)[0].Amount)
	marketType, _ := json.Marshal((*tickets)[0].MarketType)
	betData := request.BetData{
		EventID: eventID,
		TableID: tableID,
		Tickets: []request.BetTicketData{
			{
				Selection:     selectionData.Selection,
				Amount:        amount,
				MarketType:    marketType,
				SelectionData: types.JSONRaw((*tickets)[0].SelectionData),
			},
		},
	}
	if prevMemberLevel := bp.GetPrevMemberLevel(); prevMemberLevel == nil { //no active ticket
		return response.ErrorIE(errors.VALIDATE_SELECTION_ERROR, errors.IEID_INVALID_SELECTION, process_bet.SelectionType)
	}
	if selectionData.Selection.String() == constants_loltower.SelectionSkip && (bp.prevMemberLevel.Skip-1) < 0 {
		return response.ErrorIE(errors.SKIP_LIMIT_REACHED, errors.IEID_SKIP_LIMIT_REACHED, process_bet.SelectionType)
	}
	if err := bp.OverrideBetDataIfNeeded(&betData); err != nil {
		return response.ErrorIE(errors.VALIDATE_SELECTION_ERROR, errors.IEID_INVALID_SELECTION, process_bet.SelectionType)
	}
	return bp.processDatasource.PrepareBet(&betData)
}

func (bp *betProcess) ProcessPayoutSelectionIfNeeded(selectionData *request.SelectionData) response.ResponseData {
	if selectionData.Selection.String() != constants_loltower.SelectionPayout {
		return nil
	}
	if err := bp.processDatasource.Payout(constants_loltower.SelectionPayout); err != nil {
		return err
	}
	bet := response.BetData{
		UserID: bp.processDatasource.GetEncryptedUserID(),
	}
	if betTickets := bp.processDatasource.GetBetTickets(); betTickets != nil {
		bet.Tickets = betTickets
		if tickets := bp.processDatasource.GetTickets(); tickets != nil {
			(*bet.Tickets)[0].Status = constants.TICKET_STATUS_SETTLED_PENDING_PAYOUT
			(*bet.Tickets)[0].Selection = constants_loltower.SelectionPayout
			(*bet.Tickets)[0].WinLossAmount = *types.Float(*(*tickets)[0].WinLossAmount).FixedStr(2).Ptr()
			(*bet.Tickets)[0].PayoutAmount = *types.Float(*(*tickets)[0].PayoutAmount).FixedStr(2).Ptr()
		}
	}
	go bp.processDatasource.CheckDailyWinLoss()
	return &bet
}

func (bp *betProcess) CleanUp() {
	bp.processDatasource.CleanUp()
	bp.walletProcess.CleanUp()
	bp.prevMemberLevel = nil
}

// process_wallet datasource
func (bp *betProcess) GetIdentifier() string {
	return bp.datasource.GetIdentifier()
}

func (bp *betProcess) GetUser() *models.User {
	return bp.datasource.GetUser()
}

func (bp *betProcess) GetTransactions() *[]process_wallet.TRequest {
	user := bp.datasource.GetUser()
	event := bp.datasource.GetEvent()
	tickets := bp.processDatasource.GetTickets()
	comboTickets := bp.processDatasource.GetComboTickets()
	eventID := (*string)(nil)
	eventName := *types.Int(*event.ID).String().Ptr()
	transactions := []process_wallet.TRequest{}

	if event.ESID != nil {
		eventID = types.Int(*event.ESID).String().Ptr()
	}
	if settings.GetEnvironment().String() == "local" {
		eventID = types.Int(*event.ID).String().Ptr()
	}
	for i := 0; i < len(*tickets); i++ {
		odds := types.Odds((*tickets)[i].Odds)
		oddsStr := odds.String(2)
		euroOdds := types.Odds((*tickets)[i].EuroOdds)
		euroOddsStr := euroOdds.String(2)
		malayOddsStr := euroOdds.EuroToMalay(4).String()
		transactions = append(transactions, process_wallet.TRequest{
			TransactionType: process_wallet.TransactionTypeBet,
			Amount:          -1 * (*tickets)[i].Amount,
			RefNo:           uuid.NewString(),
			MemberID:        user.EsportsID,
			AutoRollback:    true,
			TicketDetails: &process_wallet.TRequestTicket{
				ID:               (*tickets)[i].ID,
				Odds:             euroOddsStr,
				Amount:           (*tickets)[i].Amount,
				Currency:         user.CurrencyCode,
				Earnings:         utils.Ptr(0.0),
				EventID:          eventID,
				IsCombo:          true,
				EuroOdds:         euroOddsStr,
				EventName:        eventName,
				MemberCode:       user.MemberCode,
				MemberOdds:       euroOddsStr,
				TicketType:       constants_loltower.TicketType,
				DateCreated:      utils.TimeNow(),
				ModifiedDateTime: utils.Ptr(utils.TimeNow()),
				GameTypeID:       constants_loltower.ESGameID,
				IsUnsettled:      false,
				EventDatetime:    event.Ctime,
				GameTypeName:     constants_loltower.GameTypeName,
				RequestSource:    (*tickets)[i].RequestSource,
				CompetitionName:  constants_loltower.CompetitionName,
				MemberOddsStyle:  constants_loltower.MemberOddsStyle,
				SettlementStatus: constants.SETTLEMENT_STATUS_CONFIRMED,
				Tickets: &[]process_wallet.TRequestTicket{{
					ID:               (*comboTickets)[i].ID,
					Odds:             oddsStr,
					Currency:         user.CurrencyCode,
					EventID:          eventID,
					EuroOdds:         euroOddsStr,
					EventName:        eventName,
					MalayOdds:        &malayOddsStr,
					MemberOdds:       euroOddsStr,
					TicketType:       constants_loltower.TicketType,
					DateCreated:      utils.TimeNow(),
					GameTypeID:       constants_loltower.ESGameID,
					IsUnsettled:      false,
					BetSelection:     (*comboTickets)[i].Selection,
					BetTypeName:      types.String(constants_loltower.BetTypeName).Ptr(),
					MarketOption:     types.String(constants_loltower.MarketOption).Ptr(),
					EventDatetime:    event.Ctime,
					GameTypeName:     constants_loltower.GameTypeName,
					RequestSource:    (*comboTickets)[i].RequestSource,
					CompetitionName:  constants_loltower.CompetitionName,
					GameMarketName:   constants_loltower.GameMarketName,
					MemberOddsStyle:  constants_loltower.MemberOddsStyle,
					SettlementStatus: constants.SETTLEMENT_STATUS_CONFIRMED,
				}},
			},
		})
	}
	return &transactions
}

// override
func (bp *betProcess) OverrideBetDataIfNeeded(betData *request.BetData) error {
	//TESTING START
	if types.Array[string](strings.Split(constants_loltower.SelectionWinLoseEnabledEnv, ",")).Constains(settings.GetEnvironment().String()) {
		if betData.EventID.String() == "auto" && bp.datasource.GetEvent() != nil {
			if eventIDBytes, err := json.Marshal(bp.datasource.GetEvent().ID); err != nil {
				logger.Error("OverrideBetDataIfNeeded EventID == auto error: ", err.Error())
			} else {
				betData.EventID = eventIDBytes
			}
		}
		selection := betData.Tickets[0].Selection.String()

		if (types.Array[string]{constants_loltower.SelectionWin, constants_loltower.SelectionLose}).Constains(selection) {
			eventResult := (*bp.datasource.GetEventResults())[0]
			results := strings.Split(eventResult.Value, ",")

			if selection == constants_loltower.SelectionLose {
				selections := []string{
					*constants_loltower.Selection1.String().Ptr(),
					*constants_loltower.Selection2.String().Ptr(),
					*constants_loltower.Selection3.String().Ptr(),
					*constants_loltower.Selection4.String().Ptr(),
					*constants_loltower.Selection5.String().Ptr(),
				}
				invResults := []string{}

				for i := 0; i < len(selections); i++ {
					if !(types.Array[string](results)).Constains(selections[i]) {
						invResults = append(invResults, selections[i])
					}
				}
				results = invResults
			}
			rSource := rand.NewSource(time.Now().UnixNano())
			random := rand.New(rSource)
			randomSelection := results[random.Int()%len(results)]

			if selectionBytes, err := json.Marshal(randomSelection); err != nil {
				logger.Error("OverrideBetDataIfNeeded Tickets.Selection ", selection, " error:", err.Error())
			} else {
				betData.Tickets[0].Selection = selectionBytes
			}
		}
	}
	//TESTING END
	level := int8(1)
	euroOdds := OddsFromLevel(level)
	maxPayoutEuroOdds := euroOdds

	if prevMemberLevel := bp.GetPrevMemberLevel(); prevMemberLevel != nil {
		if betData.Tickets[0].Selection.String() == constants_loltower.SelectionSkip {
			level = prevMemberLevel.Level
		} else {
			level = prevMemberLevel.Level + 1
		}
		maxPayoutEuroOdds = OddsFromLevel(level)
		euroOdds = utils.IfElse(level == 1, maxPayoutEuroOdds, maxPayoutEuroOdds/OddsFromLevel(level-1))
	}
	betData.Tickets[0].EuroOdds = *euroOdds.Round(2).Ptr()
	betData.Tickets[0].MalayOdds = *euroOdds.EuroToMalay(4).Ptr()
	betData.Tickets[0].HongkongOdds = *euroOdds.EuroToHK(2).Ptr()
	betData.Tickets[0].MaxPayoutEuroOdds = *maxPayoutEuroOdds.Ptr()
	return nil
}

// callbacks
func (bp *betProcess) BetOpenRangeCallback() *process_bet.BetOpenRange {
	return &process_bet.BetOpenRange{
		MinMS: 0,
		MaxMS: constants_loltower.StartBetMS + constants_loltower.StopBetMS,
	}
}

func (bp *betProcess) BetResultCallback(betTicket request.BetTicketData) *process_bet.BetResult {
	betAmount := betTicket.Amount.ToFloat64()
	betResult := process_bet.BetResult{
		Result:                 1,
		WinLossAmount:          -betAmount,
		PayoutAmount:           0,
		HongkongOdds:           betTicket.HongkongOdds,
		EuroOdds:               betTicket.EuroOdds,
		MalayOdds:              betTicket.MalayOdds,
		OriginalOdds:           betTicket.HongkongOdds,
		PossibleWinningsAmount: betTicket.EuroOdds * betAmount,
	}

	if prevMemberLevel := bp.GetPrevMemberLevel(); prevMemberLevel != nil {
		totalEuroOdds := OddsFromLevel(utils.IfElse(betTicket.Selection.String() == constants_loltower.SelectionSkip, prevMemberLevel.Level, prevMemberLevel.Level+1))

		betResult.PossibleWinningsAmount = *totalEuroOdds.Precision(2).Ptr() * betAmount
	}
	if betTicket.Selection.String() == constants_loltower.SelectionSkip {
		winLossAmount := betTicket.HongkongOdds * betAmount
		payoutAmount := betAmount + winLossAmount

		betResult.Result = 0
		betResult.WinLossAmount = winLossAmount
		betResult.PayoutAmount = payoutAmount
	} else if eventResults := bp.datasource.GetEventResults(); eventResults != nil {
		if types.Array[string](strings.Split((*eventResults)[0].Value, ",")).Constains(betTicket.Selection.String()) {
			winLossAmount := betTicket.HongkongOdds * betAmount
			payoutAmount := betAmount + winLossAmount

			betResult.Result = 0
			betResult.WinLossAmount = winLossAmount
			betResult.PayoutAmount = payoutAmount
		}
	}
	return &betResult
}
