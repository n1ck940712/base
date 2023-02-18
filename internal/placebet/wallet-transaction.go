package placebet

import (
	"time"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/api"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/constants"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/errors"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/settings"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/types"
	"github.com/google/uuid"
)

func (p *Placebet) initWalletTransaction() (WTResponse, errors.FinalErrorMessage) {
	ticketDetails := p.prepareWalletTransaction()
	wtRequest := WTCRequest{
		MemberID:        p.userMgDetails.EsportsID,
		Amount:          p.ticket.Amount * -1,
		TransactionType: constants.TRANSACTION_TYPE_BET,
		RefNo:           uuid.NewString(),
		AutoRollback:    true,
		TicketDetail:    ticketDetails,
	}
	wtResponse := WTResponse{}
	err := api.NewAPI(settings.EBO_API + "/v4/wallet/").
		SetIdentifier("initWalletTransaction").
		AddHeaders(map[string]string{
			"User-Agent":    settings.USER_AGENT,
			"Authorization": settings.SERVER_TOKEN,
			"Content-Type":  "application/json",
		}).
		AddBody(wtRequest).
		Post(&wtResponse)

	if err != nil {
		logger.Error("Error initWalletTransaction", err.Error(), " response: ", wtResponse)
		return wtResponse, errors.FinalizeErrorMessage(errors.WALLET_ERROR, errors.IEID_WALLET_ERROR, false)
	}
	return wtResponse, nil
}

func (p *Placebet) commitTransaction(initWTResponse WTResponse) (WTResponse, errors.FinalErrorMessage) {
	wtResponse := WTResponse{}
	err := api.NewAPI(settings.EBO_API + "/v4/wallet/" + initWTResponse.ID + "/commit/").
		SetIdentifier("commitTransaction").
		AddHeaders(map[string]string{
			"User-Agent":    settings.USER_AGENT,
			"Authorization": settings.SERVER_TOKEN,
			"Content-Type":  "application/json",
		}).
		Post(wtResponse)

	if err != nil {
		logger.Error("Error commitTransaction", err.Error(), " response: ", wtResponse)
		return wtResponse, errors.FinalizeErrorMessage(errors.WALLET_ERROR, errors.IEID_WALLET_ERROR_COMMIT, false)
	}
	return wtResponse, nil
}

func (p *Placebet) rollbackTransaction(initWTResponse WTResponse) (WTResponse, errors.FinalErrorMessage) {
	wtResponse := WTResponse{}
	err := api.NewAPI(settings.EBO_API + "/v4/wallet/" + initWTResponse.ID + "/rollback/").
		SetIdentifier("rollbackTransaction").
		AddHeaders(map[string]string{
			"User-Agent":    settings.USER_AGENT,
			"Authorization": settings.SERVER_TOKEN,
			"Content-Type":  "application/json",
		}).
		AddBody(WTCRequest{RefNo: uuid.NewString(), MemberID: initWTResponse.MemberID}).
		Post(wtResponse)

	if err != nil {
		logger.Error("Error rollbackTransaction", err.Error(), " response: ", wtResponse)
		return wtResponse, errors.FinalizeErrorMessage(errors.WALLET_ERROR, errors.IEID_WALLET_ERROR_ROLLBACK, false)
	}
	return wtResponse, nil
}

func (p *Placebet) prepareWalletTransaction() WTRCTicket {
	// SPI bet transaction > qa will place bet to get a sample for ticket details
	timeNow := time.Now()
	earnings := 0.0
	tEventID := convertInt64ToString(p.event.ESID)
	ctEventID := convertInt64ToString(p.event.ESID)
	if settings.ENVIRONMENT == "local" {
		tEventID = convertInt64ToString(&p.ticket.EventID)
		ctEventID = convertInt64ToString(&p.comboTicket.EventID)
	}
	ctMalayOdds := types.Odds(p.comboTicket.EuroOdds).EuroToMalay(4)
	ctBetTypesName := constants.BET_TYPE_SPOR
	ctBetMarketOption := constants.LOL_TOWER_MARKET_OPTION
	walletTransaction := WTRCTicket{
		ID:                 p.ticket.ID,
		Odds:               *types.Float(p.ticket.EuroOdds).FixedStr(2).Ptr(),
		Amount:             p.ticket.Amount,
		Currency:           p.userMgDetails.CurrencyCode,
		Earnings:           &earnings,
		EventID:            tEventID,
		Handicap:           nil,
		IsCombo:            true,
		EuroOdds:           *types.Float(p.ticket.EuroOdds).FixedStr(2).Ptr(),
		EventName:          *types.Int(p.ticket.EventID).String().Ptr(),
		MalayOdds:          nil,
		MemberCode:         p.userMgDetails.MemberCode,
		MemberOdds:         *types.Float(p.ticket.EuroOdds).FixedStr(2).Ptr(),
		TicketType:         constants.MG_DEFAULT_DB_STATUS,
		DateCreated:        timeNow,
		GameTypeID:         constants.LOL_TOWER_ES_GAME_ID,
		IsUnsettled:        false,
		BetTypeName:        nil,
		MarketOption:       nil,
		ResultStatus:       nil,
		EventDatetime:      p.event.Ctime,
		GameTypeName:       constants.LOL_TOWER_GAME_NAME,
		RequestSource:      p.ticket.RequestSource,
		CompetitionName:    constants.LOL_TOWER_COMPETITION,
		MemberOddsStyle:    "euro",
		ModifiedDateTime:   &timeNow,
		SettlementStatus:   constants.SETTLEMENT_STATUS_CONFIRMED,
		SettlementDateTime: nil,
		Tickets: []WTRCTicket{
			{
				ID:                 p.comboTicket.ID,
				Odds:               *types.Float(p.comboTicket.Odds).FixedStr(2).Ptr(),
				Result:             nil,
				MapNum:             nil,
				Currency:           p.userMgDetails.CurrencyCode,
				Earnings:           nil,
				EventID:            ctEventID,
				Handicap:           nil,
				EuroOdds:           *types.Float(p.comboTicket.EuroOdds).FixedStr(2).Ptr(),
				EventName:          *types.Int(p.comboTicket.EventID).String().Ptr(),
				MalayOdds:          types.Float(ctMalayOdds).FixedStr(2).Ptr(),
				MemberOdds:         *types.Float(p.comboTicket.Odds).FixedStr(2).Ptr(),
				TicketType:         constants.MG_DEFAULT_DB_STATUS,
				DateCreated:        timeNow,
				GameTypeID:         constants.LOL_TOWER_ES_GAME_ID,
				IsUnsettled:        false,
				BetSelection:       p.comboTicket.Selection,
				BetTypeName:        &ctBetTypesName,
				MarketOption:       &ctBetMarketOption,
				EventDatetime:      p.event.Ctime,
				GameTypeName:       constants.LOL_TOWER_GAME_NAME,
				RequestSource:      p.comboTicket.RequestSource,
				CompetitionName:    constants.LOL_TOWER_COMPETITION,
				GameMarketName:     constants.GAME_MARKET_NAME[constants.LOL_TOWER_GAME_ID],
				MemberOddsStyle:    constants.DEFAULT_MEMBER_ODDS_STYLE,
				ModifiedDateTime:   nil,
				SettlementStatus:   constants.SETTLEMENT_STATUS_CONFIRMED, // to check for the correct value
				SettlementDateTime: nil,
			},
		},
	}

	return walletTransaction
}
