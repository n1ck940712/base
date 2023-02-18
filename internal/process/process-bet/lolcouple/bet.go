package process_bet

import (
	"encoding/json"
	"strings"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/constants"
	constants_lolcouple "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/lolcouple"
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
	bp.OverrideBetDataIfNeeded(betData)
	if err := bp.processDatasource.PrepareBet(betData); err != nil {
		return err
	}
	if err := bp.processDatasource.PreCreateTicketValidation(); err != nil {
		return err
	}
	if err := bp.processDatasource.PrepareTickets(); err != nil {
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
	if err := bp.processDatasource.CreateTickets(func(tx *gorm.DB) error {
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
	if betTickets := bp.processDatasource.GetBetTickets(); betTickets != nil {
		bet.Tickets = betTickets
	}
	go bp.processDatasource.CheckDailyWinLoss()
	return &bet
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
		euroOdds := types.Odds((*tickets)[i].EuroOdds)
		euroOddsStr := euroOdds.String(2)
		malayOddsStr := euroOdds.EuroToMalay(4).String()
		hkOddsStr := euroOdds.EuroToHK(2).String(2)

		transactions = append(transactions, process_wallet.TRequest{
			TransactionType: process_wallet.TransactionTypeBet,
			Amount:          -1 * (*tickets)[i].Amount,
			RefNo:           uuid.NewString(),
			MemberID:        user.EsportsID,
			TicketID:        (*tickets)[i].ID,
			AutoRollback:    true,
			TicketDetails: &process_wallet.TRequestTicket{
				ID:               (*tickets)[i].ID,
				BetSelection:     (*tickets)[i].Selection,
				Odds:             malayOddsStr,
				Currency:         user.CurrencyCode,
				Amount:           (*tickets)[i].Amount,
				GameTypeName:     constants_lolcouple.GameTypeName,
				GameMarketName:   constants_lolcouple.GameMarketName,
				MarketOption:     types.String(constants_lolcouple.MarketOption).Ptr(), //next map_num nil
				BetTypeName:      types.String(constants_lolcouple.BetTypeName).Ptr(),
				CompetitionName:  constants_lolcouple.CompetitionName,
				EventID:          eventID,
				EventName:        eventName,
				EventDatetime:    event.Ctime,
				DateCreated:      utils.TimeNow(), //next settlement_datetime nil
				ModifiedDateTime: utils.Ptr(utils.TimeNow()),
				SettlementStatus: constants.SETTLEMENT_STATUS_CONFIRMED, //next result nil; resut_status nil;
				Earnings:         types.Float(0.0).Ptr(),                //next handicap nil
				IsCombo:          false,
				MemberCode:       user.MemberCode,
				IsUnsettled:      false,
				MalayOdds:        &malayOddsStr,
				EuroOdds:         euroOddsStr,
				MemberOdds:       hkOddsStr,
				MemberOddsStyle:  constants_lolcouple.MemberOddsStyle,
				GameTypeID:       constants_lolcouple.ESGameID,
				RequestSource:    (*tickets)[i].RequestSource,
				TicketType:       constants_lolcouple.TicketType,
			},
		})
	}

	return &transactions

}

// override
func (bp *betProcess) OverrideBetDataIfNeeded(betData *request.BetData) {
	//add odds based on selection
	for i := 0; i < len(betData.Tickets); i++ {
		selection := betData.Tickets[i].Selection.String()
		euroOdds := GetSelectionOdds(selection, false)

		betData.Tickets[i].EuroOdds = *euroOdds.Ptr()
		betData.Tickets[i].MalayOdds = *euroOdds.EuroToMalay(4).Ptr()
		betData.Tickets[i].HongkongOdds = *euroOdds.EuroToHK(2).Ptr()
		if selection == constants_lolcouple.Selection10 {
			betData.Tickets[i].MaxPayoutEuroOdds = 1 / 0.03
		} else {
			betData.Tickets[i].MaxPayoutEuroOdds = *euroOdds.Ptr()
		}
	}
	//testing
	if types.Array[string](strings.Split(constants_lolcouple.SelectionWinLoseEnabledEnv, ",")).Constains(settings.GetEnvironment().String()) {
		if betData.EventID.String() == "auto" && bp.datasource.GetEvent() != nil {
			if eventIDBytes, err := json.Marshal(bp.datasource.GetEvent().ID); err != nil {
				logger.Error("OverrideBetDataIfNeeded EventID == auto error: ", err.Error())
			} else {
				betData.EventID = eventIDBytes
			}
		}
	}
}

// callbacks
func (bp *betProcess) BetOpenRangeCallback() *process_bet.BetOpenRange {
	return &process_bet.BetOpenRange{
		MinMS: 0,
		MaxMS: constants_lolcouple.StartBetMS + constants_lolcouple.StopBetMS,
	}
}

func (bp *betProcess) BetResultCallback(betTicket request.BetTicketData) *process_bet.BetResult {
	eventResults := bp.datasource.GetEventResults()
	betAmount := betTicket.Amount.ToFloat64()
	betResult := process_bet.BetResult{
		Result:        1,
		WinLossAmount: -betAmount,
		PayoutAmount:  0,
		HongkongOdds:  betTicket.HongkongOdds,
		EuroOdds:      betTicket.EuroOdds,
		MalayOdds:     betTicket.MalayOdds,
		OriginalOdds:  betTicket.HongkongOdds,
	}

	if eventResults != nil {
		for i := 0; i < len(*eventResults); i++ {
			results := []string{}

			if err := json.Unmarshal([]byte((*eventResults)[i].Value), &results); err != nil {
				logger.Error(bp.datasource.GetIdentifier(), " bet BetResultCallback error: ", err.Error())
				break
			}
			if len(results) < 1 {
				logger.Error(bp.datasource.GetIdentifier(), " bet BetResultCallback error: results must contain atleast 1 element")
				break
			}
			result1 := CoupleToValue(results[0])

			if isWin := GetSelectionWin(betTicket.Selection.String(), result1); isWin {
				winLossAmount := betResult.HongkongOdds * betAmount
				payoutAmount := betAmount + winLossAmount

				betResult.Result = 0
				betResult.WinLossAmount = winLossAmount
				betResult.PayoutAmount = payoutAmount
			}

			if len(results) == 3 && result1 == constants_lolcouple.SelectionBonusCoupleValue &&
				CoupleToValue(results[1]) == constants_lolcouple.SelectionBonusCoupleValue &&
				CoupleToValue(results[2]) == constants_lolcouple.SelectionBonusCoupleValue {
				euroOdds := GetSelectionOdds(betTicket.Selection.String(), true)

				betResult.EuroOdds = *euroOdds.Ptr()
				betResult.MalayOdds = *euroOdds.EuroToMalay(4).Ptr()
				betResult.HongkongOdds = *euroOdds.EuroToHK(2).Ptr()
				winLossAmount := betResult.HongkongOdds * betAmount
				payoutAmount := betAmount + winLossAmount

				betResult.Result = 0
				betResult.WinLossAmount = winLossAmount
				betResult.PayoutAmount = payoutAmount
				break
			}
		}
	}
	return &betResult
}

func CoupleToValue(couple string) int8 {
	result := strings.Split(couple, ":")
	maleValue := types.String(result[0]).Int().Int8()
	femaleValue := types.String(result[1]).Int().Int8()

	return maleValue + femaleValue
}

func GetSelectionOdds(selection string, isBonus bool) types.Odds {
	if isBonus {
		switch selection {
		case constants_lolcouple.Selection5:
			return constants_lolcouple.Selection5BonusOdds
		case constants_lolcouple.Selection10:
			return constants_lolcouple.Selection10Odds
		default:
			return constants_lolcouple.SelectionBonusOdds
		}
	}
	switch selection {
	case constants_lolcouple.Selection1:
		return constants_lolcouple.Selection1Odds
	case constants_lolcouple.Selection2:
		return constants_lolcouple.Selection2Odds
	case constants_lolcouple.Selection3:
		return constants_lolcouple.Selection3Odds
	case constants_lolcouple.Selection4:
		return constants_lolcouple.Selection4Odds
	case constants_lolcouple.Selection5:
		return constants_lolcouple.Selection5Odds
	case constants_lolcouple.Selection6:
		return constants_lolcouple.Selection6Odds
	case constants_lolcouple.Selection7:
		return constants_lolcouple.Selection7Odds
	case constants_lolcouple.Selection8:
		return constants_lolcouple.Selection8Odds
	case constants_lolcouple.Selection9:
		return constants_lolcouple.Selection9Odds
	case constants_lolcouple.Selection10:
		return constants_lolcouple.Selection10Odds
	default:
		panic("lolcouple bet process GetSelectionOdds invalid selection: " + selection)
	}
}

func GetSelectionWin(selection string, coupleValue int8) bool {
	switch selection {
	case constants_lolcouple.Selection1:
		if coupleValue <= 3 {
			return true
		}
	case constants_lolcouple.Selection2:
		if coupleValue <= 4 {
			return true
		}
	case constants_lolcouple.Selection3:
		if coupleValue <= 5 {
			return true
		}
	case constants_lolcouple.Selection4:
		if coupleValue <= 6 {
			return true
		}
	case constants_lolcouple.Selection5:
		return coupleValue == 7
	case constants_lolcouple.Selection6:
		if coupleValue >= 8 {
			return true
		}
	case constants_lolcouple.Selection7:
		if coupleValue >= 9 {
			return true
		}
	case constants_lolcouple.Selection8:
		if coupleValue >= 10 {
			return true
		}
	case constants_lolcouple.Selection9:
		if coupleValue >= 11 {
			return true
		}
	}
	return false
}
