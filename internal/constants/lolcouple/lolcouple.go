package constants_lolcouple

import (
	"bitbucket.org/esportsph/minigame-backend-golang/internal/constants"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/types"
)

const (
	Identifier                            = "lolcouple"
	WebsocketChannel                      = "lolcouple"
	WebsocketRestartTimeout               = (StartBetMS + StopBetMS + ShowResultMaxMS) * 3 / 4
	WebsocketSlug                         = "lol-couple"
	GameName                              = "LOL Couple"
	GameID                                = 7
	TableID                               = 13
	ESGameID                              = 38
	MaxFutureHashes                       = 1
	MaxFutureEvents                       = 16
	MaxSequencePerHash                    = 50
	EventResultType                       = 30
	StartBetMS                            = 10000
	StopBetMS                             = 3000
	ShowResult1MS                         = 6000
	ShowResult1BonusMS                    = 8000
	ShowResult2MS                         = 7000
	ShowResult3MS                         = 7000
	ShowResult3BonusMS                    = 12000
	ShowResultMaxMS                       = ShowResult1BonusMS + ShowResult2MS + ShowResult3BonusMS
	GameTypeName                          = "lol-couple"
	GameMarketName                        = "Couple"
	CompetitionName                       = "Lootbox"
	MarketOption                          = "match"
	MarketType                            = 20
	BetTypeName                           = "SPOR"
	Selection1                            = "3_under"
	Selection2                            = "4_under"
	Selection3                            = "5_under"
	Selection4                            = "6_under"
	Selection5                            = "7"
	Selection6                            = "8_over"
	Selection7                            = "9_over"
	Selection8                            = "10_over"
	Selection9                            = "11_over"
	Selection10                           = "777"
	SelectionWinLoseEnabledEnv            = "local" //local,dev for testing purposes only
	Selection1Odds             types.Odds = 10.00
	Selection2Odds             types.Odds = 5.00
	Selection3Odds             types.Odds = 3.00
	Selection4Odds             types.Odds = 2.00
	Selection5Odds             types.Odds = 5.00
	Selection6Odds             types.Odds = 2.00
	Selection7Odds             types.Odds = 3.00
	Selection8Odds             types.Odds = 5.00
	Selection9Odds             types.Odds = 10.00
	Selection10Odds            types.Odds = 177.00
	SelectionBonusOdds         types.Odds = 22.00
	Selection5BonusOdds        types.Odds = 27.00
	SelectionBonusCoupleValue             = 7
	MemberOddsStyle                       = constants.OddsStyleHongkong
	MemberCodeMaskCount                   = 4
	GroundType                            = "Neutral"
	TicketType                            = constants.TicketTypeDB
)
