package process_bet

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/constants"
	constants_fifashootup "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/fifashootup"
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
		return nil
	}); err != nil {
		if err := bp.walletProcess.RollbackTransactions(); err != nil {
			return err
		}
		return err
	}
	bet := response.BetData{
		UserID: bp.processDatasource.GetEncryptedUserID(),
	}
	comboTickets := bp.processDatasource.GetBetComboTickets()

	if comboTickets != nil {
		bet.Tickets = comboTickets
	}
	go bp.processDatasource.CheckDailyWinLoss()
	return &bet
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
		totalEuroOdds := types.Odds((*comboTickets)[0].EuroOdds)

		for i := 0; i < len(*(*tickets)[0].ComboTickets); i++ {
			totalEuroOdds *= types.Odds((*(*tickets)[0].ComboTickets)[i].EuroOdds)
		}
		totalHongkongOdds := *totalEuroOdds.EuroToHK(2).Ptr()
		winLossAmount := -(*tickets)[0].Amount
		payoutAmount := 0.0

		if *(*comboTickets)[0].Result == constants.TICKET_RESULT_WIN {
			winLossAmount = totalHongkongOdds * (*tickets)[0].Amount
			payoutAmount = (*tickets)[0].Amount + winLossAmount
		}
		return tx.Updates(models.Ticket{
			ID:            (*tickets)[0].ID,
			EventID:       (*comboTickets)[0].EventID,
			Odds:          totalHongkongOdds,
			EuroOdds:      *totalEuroOdds.Precision(2).Ptr(),
			OriginalOdds:  &totalHongkongOdds,
			Result:        (*comboTickets)[0].Result,
			WinLossAmount: &winLossAmount,
			PayoutAmount:  &payoutAmount,
			Selection:     (*comboTickets)[0].Selection,
		}).Error
	}); err != nil {
		return err
	}
	bet := response.BetData{
		UserID: bp.processDatasource.GetEncryptedUserID(),
	}
	if comboTickets := bp.processDatasource.GetBetComboTickets(); comboTickets != nil {
		bet.Tickets = comboTickets
	}
	go bp.processDatasource.CheckDailyWinLoss()
	return &bet
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
	if err := bp.OverrideBetDataIfNeeded(&betData); err != nil {
		return response.ErrorIE(errors.VALIDATE_SELECTION_ERROR, errors.IEID_INVALID_SELECTION, process_bet.SelectionType)
	}
	return bp.processDatasource.PrepareBet(&betData)
}

func (bp *betProcess) ProcessPayoutSelectionIfNeeded(selectionData *request.SelectionData) response.ResponseData {
	if selectionData.Selection.String() == constants_fifashootup.SelectionPayout {
		if err := bp.processDatasource.Payout(constants_fifashootup.SelectionPayout); err != nil {
			return err
		}
		bet := response.BetData{
			UserID: bp.processDatasource.GetEncryptedUserID(),
		}
		if betTickets := bp.processDatasource.GetBetTickets(); betTickets != nil {
			bet.Tickets = betTickets
			if tickets := bp.processDatasource.GetTickets(); tickets != nil {
				(*bet.Tickets)[0].Status = constants.TICKET_STATUS_SETTLED_PENDING_PAYOUT
				(*bet.Tickets)[0].Selection = constants_fifashootup.SelectionPayout
				(*bet.Tickets)[0].WinLossAmount = *types.Float(*(*tickets)[0].WinLossAmount).FixedStr(2).Ptr()
				(*bet.Tickets)[0].PayoutAmount = *types.Float(*(*tickets)[0].PayoutAmount).FixedStr(2).Ptr()
			}
		}
		go bp.processDatasource.CheckDailyWinLoss()
		return &bet
	}
	return nil
}

func (bp *betProcess) CleanUp() {
	bp.processDatasource.CleanUp()
	bp.walletProcess.CleanUp()
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
				TicketType:       constants_fifashootup.TicketType,
				DateCreated:      utils.TimeNow(),
				ModifiedDateTime: utils.Ptr(utils.TimeNow()),
				GameTypeID:       constants_fifashootup.ESGameID,
				IsUnsettled:      false,
				EventDatetime:    event.Ctime,
				GameTypeName:     constants_fifashootup.GameTypeName,
				RequestSource:    (*tickets)[i].RequestSource,
				CompetitionName:  constants_fifashootup.CompetitionName,
				MemberOddsStyle:  constants_fifashootup.MemberOddsStyle,
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
					TicketType:       constants_fifashootup.TicketType,
					DateCreated:      utils.TimeNow(),
					GameTypeID:       constants_fifashootup.ESGameID,
					IsUnsettled:      false,
					BetSelection:     (*comboTickets)[i].Selection,
					BetTypeName:      types.String(constants_fifashootup.BetTypeName).Ptr(),
					MarketOption:     types.String(constants_fifashootup.MarketOption).Ptr(),
					EventDatetime:    event.Ctime,
					GameTypeName:     constants_fifashootup.GameTypeName,
					RequestSource:    (*comboTickets)[i].RequestSource,
					CompetitionName:  constants_fifashootup.CompetitionName,
					GameMarketName:   constants_fifashootup.GameMarketName,
					MemberOddsStyle:  constants_fifashootup.MemberOddsStyle,
					SettlementStatus: constants.SETTLEMENT_STATUS_CONFIRMED,
				}},
			},
		})
	}
	return &transactions
}

// override
func (bp *betProcess) OverrideBetDataIfNeeded(betData *request.BetData) error {
	resultValue := bp.GetEventResultValue()
	selectionsArr := strings.Split(resultValue, ",")
	leftCard := strings.Split(selectionsArr[0], "-")[0]
	rightCard := strings.Split(selectionsArr[1], "-")[0]
	resultCard := strings.Split(selectionsArr[2], "-")[0]
	leftRange, middleRange, rightRange := GenerateLMRRanges(leftCard, rightCard)
	leftOdds := GenerateOdds(leftRange)
	middleOdds := GenerateOdds(middleRange)
	rightOdds := GenerateOdds(rightRange)

	//TESTING START
	if types.Array[string](strings.Split(constants_fifashootup.SelectionWinLoseEnabledEnv, ",")).Constains(settings.GetEnvironment().String()) {
		if betData.EventID.String() == "auto" && bp.datasource.GetEvent() != nil {
			if eventIDBytes, err := json.Marshal(bp.datasource.GetEvent().ID); err != nil {
				logger.Error("OverrideBetDataIfNeeded EventID == auto error: ", err.Error())
			} else {
				betData.EventID = eventIDBytes
			}
		}
		selection := betData.Tickets[0].Selection.String()

		if (types.Array[string]{constants_fifashootup.SelectionWin, constants_fifashootup.SelectionLose}).Constains(selection) {
			results := []string{}

			if types.Array[string](leftRange).Constains(strings.ToUpper(resultCard)) {
				results = append(results, constants_fifashootup.Selection1)
			} else if types.Array[string](middleRange).Constains(strings.ToUpper(resultCard)) {
				results = append(results, constants_fifashootup.Selection2)
			} else if types.Array[string](rightRange).Constains(strings.ToUpper(resultCard)) {
				results = append(results, constants_fifashootup.Selection3)
			}
			if selection == constants_fifashootup.SelectionLose {
				selections := []string{constants_fifashootup.Selection1, constants_fifashootup.Selection3}
				invResults := []string{}

				if middleOdds > 0 {
					selections = append(selections, constants_fifashootup.Selection2)
				}
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
	euroOdds := types.Odds(0)

	switch betData.Tickets[0].Selection.String() {
	case constants_fifashootup.Selection1:
		euroOdds = leftOdds
	case constants_fifashootup.Selection2:
		if middleOdds == 0 {
			return fmt.Errorf("selection is %v but odds is %v", constants_fifashootup.Selection2, middleOdds)
		}
		euroOdds = middleOdds
	case constants_fifashootup.Selection3:
		euroOdds = rightOdds
	}
	betData.Tickets[0].EuroOdds = *euroOdds.Ptr()
	betData.Tickets[0].MalayOdds = *euroOdds.EuroToMalay(4).Ptr()
	betData.Tickets[0].HongkongOdds = *euroOdds.EuroToHK(2).Ptr()
	maxPayoutEuroOdds := types.Odds(math.Max(math.Max(*leftOdds.Ptr(), *middleOdds.Ptr()), *rightOdds.Ptr()))

	if tickets := bp.processDatasource.GetTickets(); tickets != nil {
		for i := 0; i < len(*(*tickets)[0].ComboTickets); i++ {
			maxPayoutEuroOdds *= types.Odds((*(*tickets)[0].ComboTickets)[i].EuroOdds)
		}
	}
	betData.Tickets[0].MaxPayoutEuroOdds = *maxPayoutEuroOdds.Ptr()
	return nil
}

// callback
func (bp *betProcess) BetOpenRangeCallback() *process_bet.BetOpenRange {
	return &process_bet.BetOpenRange{
		MinMS: 0,
		MaxMS: constants_fifashootup.StartBetMS + constants_fifashootup.StopBetMS,
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
	resultValue := bp.GetEventResultValue()
	selectionsArr := strings.Split(resultValue, ",")
	leftCard := strings.Split(selectionsArr[0], "-")[0]
	rightCard := strings.Split(selectionsArr[1], "-")[0]
	resultCard := strings.Split(selectionsArr[2], "-")[0]
	leftRange, middleRange, rightRange := GenerateLMRRanges(leftCard, rightCard)

	if tickets := bp.processDatasource.GetTickets(); tickets != nil && (*tickets)[0].ComboTickets != nil {
		totalEuroOdds := types.Odds(betTicket.EuroOdds)

		for i := 0; i < len(*(*tickets)[0].ComboTickets); i++ {
			totalEuroOdds *= types.Odds((*(*tickets)[0].ComboTickets)[i].EuroOdds)
		}
		betResult.PossibleWinningsAmount = *totalEuroOdds.Precision(2).Ptr() * betAmount
	}
	if types.Array[string](leftRange).Constains(strings.ToUpper(resultCard)) && betTicket.Selection.String() == constants_fifashootup.Selection1 ||
		types.Array[string](middleRange).Constains(strings.ToUpper(resultCard)) && betTicket.Selection.String() == constants_fifashootup.Selection2 ||
		types.Array[string](rightRange).Constains(strings.ToUpper(resultCard)) && betTicket.Selection.String() == constants_fifashootup.Selection3 {
		winLossAmount := betResult.HongkongOdds * betAmount
		payoutAmount := betAmount + winLossAmount

		betResult.Result = 0
		betResult.WinLossAmount = winLossAmount
		betResult.PayoutAmount = payoutAmount
	}
	return &betResult
}

func (bp *betProcess) GetEventResultValue() (value string) {
	eventResults := bp.datasource.GetEventResults()

	for i := 0; i < len(*eventResults); i++ {
		if (*eventResults)[i].ResultType == constants_fifashootup.EventResultType1 {
			value = (*eventResults)[i].Value
		}
	}
	return value
}
