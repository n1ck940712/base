package constants_fishprawncrab

import (
	"bitbucket.org/esportsph/minigame-backend-golang/internal/constants"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/types"
)

const (
	Identifier                            = "fishprawncrab"
	WebsocketChannel                      = "fishprawncrab"
	WebsocketRestartTimeout               = StartBetMS + StopBetMS + ShowResultMS
	WebsocketSlug                         = "fish-prawn-crab"
	GameName                              = "Fish Prawn Crab"
	GameID                                = 10
	TableID                               = 16
	ESGameID                              = 43
	MaxFutureHashes                       = 1
	MaxFutureEvents                       = 16
	MaxSequencePerHash                    = 100
	EventResultType                       = 34
	StartBetMS                            = 45000
	StopBetMS                             = 3000
	ShowResultMS                          = 8000
	GameTypeName                          = "fish-prawn-crab"
	GameMarketNameSingle                  = "Single"
	GameMarketNameDouble                  = "Double"
	GameMarketNameTriple                  = "Triple"
	CompetitionName                       = "Lootbox"
	MarketOption                          = "match"
	MarketTypeSingle                      = 23
	MarketTypeDouble                      = 24
	MarketTypeTriple                      = 25
	BetTypeName                           = "SPOR"
	Selection1                            = "tiger"
	Selection2                            = "ground"
	Selection3                            = "chicken"
	Selection4                            = "fish"
	Selection5                            = "prawn"
	Selection6                            = "crab"
	SelectionWin                          = "win"
	SelectionLose                         = "lose"
	SelectionWinLoseEnabledEnv            = "local,dev" //local,dev for testing purposes only
	SingleOdds                 types.Odds = 2.00
	DoubleOdds                 types.Odds = 11.00
	TripleOdds                 types.Odds = 180.00
	MemberOddsStyle                       = constants.OddsStyleHongkong
	MemberCodeMaskCount                   = 4
	GroundType                            = "Neutral"
	TicketType                            = constants.TicketTypeDB
)

func GetMarketTypeMarketName(marketType int16) string {
	switch marketType {
	case MarketTypeSingle:
		return GameMarketNameSingle
	case MarketTypeDouble:
		return GameMarketNameDouble
	case MarketTypeTriple:
		return GameMarketNameTriple
	default:
		panic("fishprawncrab GetMarketNameForMarketType marketType: " + types.Int(marketType).String() + " is not supported")
	}
}

func GetMarketTypeOdds(marketType int16) types.Odds {
	switch marketType {
	case MarketTypeSingle:
		return SingleOdds
	case MarketTypeDouble:
		return DoubleOdds
	case MarketTypeTriple:
		return TripleOdds
	default:
		panic("fishprawncrab GetMarketTypeOdds marketType: " + types.Int(marketType).String() + " is not supported")
	}
}
