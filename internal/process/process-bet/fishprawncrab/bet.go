package process_bet

import (
	"encoding/json"
	"math/rand"
	"strings"
	"time"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/constants"
	constants_fishprawncrab "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/fishprawncrab"
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
		for i := 0; i < len(*betTickets); i++ {
			euroOdds := constants_fishprawncrab.GetMarketTypeOdds((*betTickets)[i].MarketType)

			(*betTickets)[i].Odds = euroOdds.EuroToHK(2).String() //reset odds shown to FE
		}
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
				GameTypeName:     constants_fishprawncrab.GameTypeName,
				GameMarketName:   constants_fishprawncrab.GetMarketTypeMarketName((*tickets)[i].MarketType),
				MarketOption:     types.String(constants_fishprawncrab.MarketOption).Ptr(), //next map_num nil
				BetTypeName:      types.String(constants_fishprawncrab.BetTypeName).Ptr(),
				CompetitionName:  constants_fishprawncrab.CompetitionName,
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
				MemberOddsStyle:  constants_fishprawncrab.MemberOddsStyle,
				GameTypeID:       constants_fishprawncrab.ESGameID,
				RequestSource:    (*tickets)[i].RequestSource,
				TicketType:       constants_fishprawncrab.TicketType,
			},
		})
	}

	return &transactions

}

// override
func (bp *betProcess) OverrideBetDataIfNeeded(betData *request.BetData) {
	//add odds based on selection
	for i := 0; i < len(betData.Tickets); i++ {
		euroOdds := constants_fishprawncrab.GetMarketTypeOdds(betData.Tickets[i].MarketType.Int16())

		betData.Tickets[i].EuroOdds = *euroOdds.Ptr()
		betData.Tickets[i].MalayOdds = *euroOdds.EuroToMalay(4).Ptr()
		betData.Tickets[i].HongkongOdds = *euroOdds.EuroToHK(2).Ptr()
		betData.Tickets[i].MaxPayoutEuroOdds = *euroOdds.Ptr()
	}

	//testing
	if types.Array[string](strings.Split(constants_fishprawncrab.SelectionWinLoseEnabledEnv, ",")).Constains(settings.GetEnvironment().String()) {
		if betData.EventID.String() == "auto" && bp.datasource.GetEvent() != nil {
			if eventIDBytes, err := json.Marshal(bp.datasource.GetEvent().ID); err != nil {
				logger.Error("OverrideBetDataIfNeeded EventID == auto error: ", err.Error())
			} else {
				betData.EventID = eventIDBytes
			}
		}
		for i := 0; i < len(betData.Tickets); i++ {
			selection := betData.Tickets[i].Selection.String()

			if (types.Array[string]{constants_fishprawncrab.SelectionWin, constants_fishprawncrab.SelectionLose}).Constains(selection) {
				selections := []string{
					constants_fishprawncrab.Selection1,
					constants_fishprawncrab.Selection2,
					constants_fishprawncrab.Selection3,
					constants_fishprawncrab.Selection4,
					constants_fishprawncrab.Selection5,
					constants_fishprawncrab.Selection6,
				}
				results := strings.Split((*bp.datasource.GetEventResults())[0].Value, ",")
				lossResults := []string{}

				for i := 0; i < len(selections); i++ {
					if !(types.Array[string](results)).Constains(selections[i]) {
						lossResults = append(lossResults, selections[i])
					}
				}
				if selection == constants_fishprawncrab.SelectionLose {
					results = lossResults
				}
				rSource := rand.NewSource(time.Now().UnixNano())
				random := rand.New(rSource)
				randomSelection := results[random.Int()%len(results)]

				switch betData.Tickets[i].MarketType.Int() {
				case constants_fishprawncrab.MarketTypeDouble:
					occurences := map[string]int8{}

					for i := 0; i < len(results); i++ {
						occurences[results[i]] += 1
					}
					for occurence, count := range occurences {
						if count >= 2 {
							randomSelection = occurence
						}
					}
				case constants_fishprawncrab.MarketTypeTriple:
					if len(lossResults) < 5 {
						randomSelection = selections[random.Int()%len(selections)]
					}
				}
				if selectionBytes, err := json.Marshal(randomSelection); err != nil {
					logger.Error("OverrideBetDataIfNeeded MarketTypeSingle Tickets.Selection ", selection, " error:", err.Error())
				} else {
					betData.Tickets[i].Selection = selectionBytes
				}
			}
		}
	}
}

// callbacks
func (bp *betProcess) BetOpenRangeCallback() *process_bet.BetOpenRange {
	return &process_bet.BetOpenRange{
		MinMS: 0,
		MaxMS: constants_fishprawncrab.StartBetMS + constants_fishprawncrab.StopBetMS,
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
			results := strings.Split((*eventResults)[i].Value, ",")
			selectionToResultCount := CountSelectionsForResults(results, betTicket.Selection.String())
			isWin := false

			switch betTicket.MarketType.Int() {
			case constants_fishprawncrab.MarketTypeSingle:
				if selectionToResultCount > 0 {
					isWin = true
					euroOdds := constants_fishprawncrab.SingleOdds + types.Odds((selectionToResultCount - 1))

					betResult.EuroOdds = *euroOdds.Ptr()
					betResult.MalayOdds = *euroOdds.EuroToMalay(4).Ptr()
					betResult.HongkongOdds = *euroOdds.EuroToHK(2).Ptr()
					betResult.OriginalOdds = *euroOdds.EuroToHK(2).Ptr()
				}
			case constants_fishprawncrab.MarketTypeDouble:
				if selectionToResultCount >= 2 {
					isWin = true
				}
			case constants_fishprawncrab.MarketTypeTriple:
				if selectionToResultCount == 3 {
					isWin = true
				}
			}
			if isWin {
				winLossAmount := betResult.HongkongOdds * betAmount
				payoutAmount := betAmount + winLossAmount

				betResult.Result = 0
				betResult.WinLossAmount = winLossAmount
				betResult.PayoutAmount = payoutAmount
			}
		}
	}
	return &betResult
}

func CountSelectionsForResults(results []string, selection string) int8 {
	count := int8(0)

	for i := 0; i < len(results); i++ {
		if selection == results[i] {
			count += 1
		}
	}
	return count
}
