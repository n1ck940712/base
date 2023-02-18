package constants_fifashootup

import (
	"bitbucket.org/esportsph/minigame-backend-golang/internal/constants"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/types"
)

const (
	Identifier                            = "soccershootout"
	WebsocketChannel                      = "soccershootout"
	WebsocketRestartTimeout               = (StartBetMS + StopBetMS + ShowResultMS) * 3 / 4
	WebsocketSlug                         = "soccer-shootout"
	GameName                              = "Soccer Shootout"
	GameID                                = 9
	TableID                               = 15
	ESGameID                              = 40
	MaxFutureHashes                       = 1
	MaxFutureEvents                       = 16
	MaxSequencePerHash                    = 100
	EventResultType1                      = 32
	EventResultType2                      = 33
	StartBetMS                            = 10000
	StopBetMS                             = 3000
	ShowResultMS                          = 4000
	GameTypeName                          = "soccer-shootout"
	GameMarketName                        = "Shootout"
	CompetitionName                       = "Lootbox"
	MarketOption                          = "match"
	MarketType                            = 22
	BetTypeName                           = "SPOR"
	Selection1                            = "left"
	Selection2                            = "middle"
	Selection3                            = "right"
	SelectionPayout                       = "p"
	SelectionWin                          = "win"
	SelectionLose                         = "lose"
	SelectionWinLoseEnabledEnv            = "local" //local,dev for testing purposes only
	Card1Odds                  types.Odds = 12.54
	Card2Odds                  types.Odds = 6.27
	Card3Odds                  types.Odds = 4.18
	Card4Odds                  types.Odds = 3.13
	Card5Odds                  types.Odds = 2.50
	Card6Odds                  types.Odds = 2.09
	Card7Odds                  types.Odds = 1.79
	Card8Odds                  types.Odds = 1.56
	Card9Odds                  types.Odds = 1.39
	Card10Odds                 types.Odds = 1.25
	Card11Odds                 types.Odds = 1.14
	Card12Odds                 types.Odds = 1.04
	MaxBetCount                           = 5
	MemberOddsStyle                       = constants.OddsStyleHongkong
	MemberCodeMaskCount                   = 4
	GroundType                            = "Neutral"
	TicketType                            = constants.TicketTypeDB
)
