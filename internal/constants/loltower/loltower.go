package constants_loltower

import (
	"bitbucket.org/esportsph/minigame-backend-golang/internal/constants"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/types"
)

const (
	Identifier                            = "loltower"
	WebsocketChannel                      = "loltower"
	WebsocketRestartTimeout               = (StartBetMS + StopBetMS + ShowResultMS) * 3 / 4
	WebsocketSlug                         = "lol-tower"
	GameName                              = "LOL Tower"
	GameID                                = 5
	TableID                               = 11
	ESGameID                              = 33
	MaxFutureHashes                       = 1
	MaxFutureEvents                       = 16
	MaxSequencePerHash                    = 50
	EventResultType                       = 28
	MinLevel                              = 1
	MaxLevel                              = 10
	MaxSkip                               = 3
	GameDuration                          = 14
	StartBetMS                            = 7000
	StopBetMS                             = 3000
	ShowResultMS                          = 4000
	BettingDuration                       = 7
	StopBettingDuration                   = 3
	ShowResultDuration                    = 4
	GameTypeName                          = "lol-tower"
	GameMarketName                        = "Tower"
	CompetitionName                       = "Lootbox"
	MarketOption                          = "match"
	MarketType                            = 18
	BetTypeName                           = "SPOR"
	SelectionSkip                         = "s"
	SelectionPayout                       = "p"
	SelectionWin                          = "win"
	SelectionLose                         = "lose"
	SelectionWinLoseEnabledEnv            = "local" //local,dev for testing purposes only
	Selection1                 types.Int  = 1
	Selection2                 types.Int  = 2
	Selection3                 types.Int  = 3
	Selection4                 types.Int  = 4
	Selection5                 types.Int  = 5
	Level1Odds                 types.Odds = 1.61
	Level2Odds                 types.Odds = 2.68
	Level3Odds                 types.Odds = 4.48
	Level4Odds                 types.Odds = 7.41
	Level5Odds                 types.Odds = 12.21
	Level6Odds                 types.Odds = 20.42
	Level7Odds                 types.Odds = 33.88
	Level8Odds                 types.Odds = 56.65
	Level9Odds                 types.Odds = 95.43
	Level10Odds                types.Odds = 160.4
	MaxBetCount                           = 10
	MemberOddsStyle                       = constants.OddsStyleHongkong
	MemberCodeMaskCount                   = 4
	GroundType                            = "Neutral"
	TicketType                            = constants.TicketTypeDB
	ChampionIntervalMS                    = 4000
)
